// SPDX-License-Identifier: GPL-2.0
// l2radar - passive L2 neighbour monitor via TC/TCX

#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <linux/if_arp.h>
#include <linux/ipv6.h>
#include <linux/icmpv6.h>
#include <linux/in6.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

char LICENSE[] SEC("license") = "GPL";

#define MAX_IPV4 4
#define MAX_IPV6 4
#define ETH_ALEN 6
#define MAX_ENTRIES 4096
#define MAX_NDP_OPTIONS 4

/* ARP opcodes */
#define ARPOP_REQUEST 1
#define ARPOP_REPLY   2

/* ICMPv6/NDP message types */
#define ICMPV6_ROUTER_SOLICITATION    133
#define ICMPV6_ROUTER_ADVERTISEMENT   134
#define ICMPV6_NEIGHBOUR_SOLICITATION 135
#define ICMPV6_NEIGHBOUR_ADVERTISEMENT 136

/* NDP option types */
#define NDP_OPT_SOURCE_LL_ADDR 1
#define NDP_OPT_TARGET_LL_ADDR 2

/* Map key: MAC address padded to 8 bytes for alignment */
struct mac_key {
	__u8 addr[ETH_ALEN];
	__u8 _pad[2];
};

/* Map value: associated IPs and timestamps */
struct neighbour_entry {
	__be32 ipv4[MAX_IPV4];
	struct in6_addr ipv6[MAX_IPV6];
	__u8 ipv4_count;
	__u8 ipv6_count;
	__u8 _pad[6];
	__u64 first_seen;
	__u64 last_seen;
};

/* ARP header for IPv4 over Ethernet (28 bytes) */
struct arp_ipv4 {
	__be16 ar_hrd;    /* hardware type */
	__be16 ar_pro;    /* protocol type */
	__u8   ar_hln;    /* hardware address length */
	__u8   ar_pln;    /* protocol address length */
	__be16 ar_op;     /* opcode */
	__u8   ar_sha[ETH_ALEN]; /* sender hardware address */
	__be32 ar_sip;    /* sender IP */
	__u8   ar_tha[ETH_ALEN]; /* target hardware address */
	__be32 ar_tip;    /* target IP */
} __attribute__((packed));

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__type(key, struct mac_key);
	__type(value, struct neighbour_entry);
	__uint(max_entries, MAX_ENTRIES);
} neighbours SEC(".maps");

/* Check if a MAC address is multicast (bit 0 of first byte set). */
static __always_inline int is_multicast(__u8 *mac)
{
	return mac[0] & 0x01;
}

/* Check if a MAC address is broadcast (ff:ff:ff:ff:ff:ff). */
static __always_inline int is_broadcast(__u8 *mac)
{
	return (mac[0] & mac[1] & mac[2] & mac[3] & mac[4] & mac[5]) == 0xff;
}

/*
 * Ensure a MAC entry exists in the map and return a pointer to it.
 * Sets first_seen on creation, updates last_seen always.
 */
static __always_inline struct neighbour_entry *track_mac(__u8 *mac)
{
	struct mac_key key = {};
	__builtin_memcpy(key.addr, mac, ETH_ALEN);

	__u64 now = bpf_ktime_get_ns();

	struct neighbour_entry *entry = bpf_map_lookup_elem(&neighbours, &key);
	if (entry) {
		entry->last_seen = now;
		return entry;
	}

	/* New entry */
	struct neighbour_entry new_entry = {};
	new_entry.first_seen = now;
	new_entry.last_seen = now;
	bpf_map_update_elem(&neighbours, &key, &new_entry, BPF_NOEXIST);

	return bpf_map_lookup_elem(&neighbours, &key);
}

/* ICMPv6 header (first 4 bytes) */
struct icmp6hdr_minimal {
	__u8 type;
	__u8 code;
	__be16 checksum;
};

/* NDP NS/NA body (after ICMPv6 header): 4 reserved/flags + 16 target */
struct ndp_ns_na {
	__u8  flags_reserved[4];
	struct in6_addr target;
};

/* NDP option header */
struct ndp_opt_hdr {
	__u8 type;
	__u8 length; /* in units of 8 bytes */
};

/* Compare two in6_addr for equality. */
static __always_inline int in6_addr_equal(const struct in6_addr *a,
					  const struct in6_addr *b)
{
	return a->in6_u.u6_addr32[0] == b->in6_u.u6_addr32[0] &&
	       a->in6_u.u6_addr32[1] == b->in6_u.u6_addr32[1] &&
	       a->in6_u.u6_addr32[2] == b->in6_u.u6_addr32[2] &&
	       a->in6_u.u6_addr32[3] == b->in6_u.u6_addr32[3];
}

/* Check if an in6_addr is all zeros. */
static __always_inline int in6_addr_is_zero(const struct in6_addr *a)
{
	return (a->in6_u.u6_addr32[0] | a->in6_u.u6_addr32[1] |
		a->in6_u.u6_addr32[2] | a->in6_u.u6_addr32[3]) == 0;
}

/*
 * Add an IPv6 address to a neighbour entry, deduplicating.
 * Respects the cap of MAX_IPV6.
 */
static __always_inline void add_ipv6(struct neighbour_entry *entry,
				     const struct in6_addr *ip)
{
	if (in6_addr_is_zero(ip))
		return;

	/* Check for duplicates */
	#pragma unroll
	for (int i = 0; i < MAX_IPV6; i++) {
		if (i >= entry->ipv6_count)
			break;
		if (in6_addr_equal(&entry->ipv6[i], ip))
			return;
	}

	/* Add if not at capacity */
	if (entry->ipv6_count < MAX_IPV6) {
		__builtin_memcpy(&entry->ipv6[entry->ipv6_count], ip,
				 sizeof(struct in6_addr));
		entry->ipv6_count++;
	}
}

/*
 * Add an IPv4 address to a neighbour entry, deduplicating.
 * Respects the cap of MAX_IPV4.
 */
static __always_inline void add_ipv4(struct neighbour_entry *entry, __be32 ip)
{
	if (ip == 0)
		return;

	/* Check for duplicates */
	#pragma unroll
	for (int i = 0; i < MAX_IPV4; i++) {
		if (i >= entry->ipv4_count)
			break;
		if (entry->ipv4[i] == ip)
			return;
	}

	/* Add if not at capacity */
	if (entry->ipv4_count < MAX_IPV4) {
		entry->ipv4[entry->ipv4_count] = ip;
		entry->ipv4_count++;
	}
}

/*
 * Process an ARP packet. Extract sender (and target for replies) MAC+IP.
 */
static __always_inline void handle_arp(void *data, void *data_end,
				       void *l3_start)
{
	struct arp_ipv4 *arp = l3_start;
	if ((void *)(arp + 1) > data_end)
		return;

	/* Validate: Ethernet + IPv4 ARP */
	if (bpf_ntohs(arp->ar_hrd) != 1 ||     /* ARPHRD_ETHER */
	    bpf_ntohs(arp->ar_pro) != 0x0800 || /* ETH_P_IP */
	    arp->ar_hln != ETH_ALEN ||
	    arp->ar_pln != 4)
		return;

	__u16 opcode = bpf_ntohs(arp->ar_op);

	/* Always process sender if unicast */
	if (!is_multicast(arp->ar_sha) && !is_broadcast(arp->ar_sha)) {
		struct neighbour_entry *entry = track_mac(arp->ar_sha);
		if (entry)
			add_ipv4(entry, arp->ar_sip);
	}

	/* For replies, also process target */
	if (opcode == ARPOP_REPLY) {
		if (!is_multicast(arp->ar_tha) && !is_broadcast(arp->ar_tha)) {
			struct neighbour_entry *entry = track_mac(arp->ar_tha);
			if (entry)
				add_ipv4(entry, arp->ar_tip);
		}
	}
}

/*
 * Associate a link-layer address from an NDP option with an IPv6 address.
 * The MAC from the option is tracked and the IPv6 address is added to it.
 */
static __always_inline void ndp_associate_ll(void *data_end,
					     struct ndp_opt_hdr *opt,
					     const struct in6_addr *ip)
{
	/* Option must be at least 8 bytes (length=1) to contain a MAC */
	if (opt->length < 1)
		return;

	__u8 *ll_addr = (__u8 *)(opt + 1);
	if ((void *)(ll_addr + ETH_ALEN) > data_end)
		return;

	if (is_multicast(ll_addr) || is_broadcast(ll_addr))
		return;

	struct neighbour_entry *entry = track_mac(ll_addr);
	if (entry)
		add_ipv6(entry, ip);
}

/*
 * Parse NDP options starting at opt_start.
 * src_ip: IPv6 address to associate with Source LL options.
 * na_target: if non-NULL, IPv6 address to associate with Target LL options.
 */
static __always_inline void parse_ndp_options(void *data_end, void *opt_start,
					      const struct in6_addr *src_ip,
					      const struct in6_addr *na_target)
{
	void *opt_ptr = opt_start;

	#pragma unroll
	for (int i = 0; i < MAX_NDP_OPTIONS; i++) {
		struct ndp_opt_hdr *opt = opt_ptr;
		if ((void *)(opt + 1) > data_end)
			break;
		if (opt->length == 0)
			break;

		__u16 opt_len = (__u16)opt->length * 8;

		if (opt->type == NDP_OPT_SOURCE_LL_ADDR) {
			ndp_associate_ll(data_end, opt, src_ip);
		} else if (opt->type == NDP_OPT_TARGET_LL_ADDR && na_target) {
			ndp_associate_ll(data_end, opt, na_target);
		}

		opt_ptr += opt_len;
		if (opt_ptr > data_end)
			break;
	}
}

/*
 * Process NDP packets (NS, NA, RS, RA).
 * Extract link-layer addresses from NDP options and associate with IPv6.
 */
static __always_inline void handle_ndp(void *data, void *data_end,
				       void *l3_start)
{
	struct ipv6hdr *ip6 = l3_start;
	if ((void *)(ip6 + 1) > data_end)
		return;

	/* Only handle ICMPv6 */
	if (ip6->nexthdr != 58) /* IPPROTO_ICMPV6 */
		return;

	void *icmp_start = (void *)(ip6 + 1);
	struct icmp6hdr_minimal *icmp = icmp_start;
	if ((void *)(icmp + 1) > data_end)
		return;

	__u8 icmp_type = icmp->type;
	void *opt_start;
	struct in6_addr *na_target = NULL;

	switch (icmp_type) {
	case ICMPV6_NEIGHBOUR_SOLICITATION:
	case ICMPV6_NEIGHBOUR_ADVERTISEMENT: {
		/* NS/NA: 4-byte ICMPv6 hdr + 4 flags/reserved + 16 target */
		struct ndp_ns_na *ndp = (struct ndp_ns_na *)((void *)icmp + 4);
		if ((void *)(ndp + 1) > data_end)
			return;
		opt_start = (void *)(ndp + 1);
		if (icmp_type == ICMPV6_NEIGHBOUR_ADVERTISEMENT)
			na_target = &ndp->target;
		break;
	}
	case ICMPV6_ROUTER_SOLICITATION:
		/* RS: 4-byte ICMPv6 hdr + 4 reserved, then options */
		opt_start = (void *)icmp + 8;
		if (opt_start > data_end)
			return;
		break;
	case ICMPV6_ROUTER_ADVERTISEMENT:
		/* RA: 4-byte ICMPv6 hdr + 12 bytes (hop limit, flags,
		 * lifetime, reachable time, retrans timer), then options */
		opt_start = (void *)icmp + 16;
		if (opt_start > data_end)
			return;
		break;
	default:
		return;
	}

	parse_ndp_options(data_end, opt_start, &ip6->saddr, na_target);
}

SEC("tc")
int l2radar(struct __sk_buff *skb)
{
	void *data = (void *)(long)skb->data;
	void *data_end = (void *)(long)skb->data_end;

	struct ethhdr *eth = data;
	if ((void *)(eth + 1) > data_end)
		return TC_ACT_UNSPEC;

	__u8 *src_mac = eth->h_source;

	/* Skip multicast and broadcast source MACs */
	if (is_multicast(src_mac) || is_broadcast(src_mac))
		return TC_ACT_UNSPEC;

	__u16 eth_proto = bpf_ntohs(eth->h_proto);
	void *l3_start = (void *)(eth + 1);

	/* Handle 802.1Q VLAN-tagged frames */
	if (eth_proto == ETH_P_8021Q) {
		/* VLAN tag: 2 bytes TCI + 2 bytes inner ethertype */
		if (l3_start + 4 > data_end)
			return TC_ACT_UNSPEC;
		eth_proto = bpf_ntohs(*(__be16 *)(l3_start + 2));
		l3_start += 4;
	}

	/*
	 * Only track MACs from frames with known ethertypes.
	 * WiFi drivers can present control/management frames with
	 * synthetic source MACs that are not real neighbours.
	 */
	switch (eth_proto) {
	case ETH_P_IP:
	case ETH_P_IPV6:
	case ETH_P_ARP:
		break;
	default:
		return TC_ACT_UNSPEC;
	}

	/* Track this unicast MAC */
	track_mac(src_mac);

	if (eth_proto == ETH_P_ARP)
		handle_arp(data, data_end, l3_start);
	else if (eth_proto == ETH_P_IPV6)
		handle_ndp(data, data_end, l3_start);

	return TC_ACT_UNSPEC;
}
