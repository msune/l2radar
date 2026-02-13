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

/* ARP opcodes */
#define ARPOP_REQUEST 1
#define ARPOP_REPLY   2

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

	/* Track this unicast MAC */
	track_mac(src_mac);

	__u16 eth_proto = bpf_ntohs(eth->h_proto);
	void *l3_start = (void *)(eth + 1);

	if (eth_proto == ETH_P_ARP)
		handle_arp(data, data_end, l3_start);

	return TC_ACT_UNSPEC;
}
