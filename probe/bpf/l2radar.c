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
 * Upsert a MAC address into the neighbours map.
 * Sets first_seen on creation, updates last_seen always.
 */
static __always_inline void track_mac(__u8 *mac)
{
	struct mac_key key = {};
	__builtin_memcpy(key.addr, mac, ETH_ALEN);

	__u64 now = bpf_ktime_get_ns();

	struct neighbour_entry *entry = bpf_map_lookup_elem(&neighbours, &key);
	if (entry) {
		entry->last_seen = now;
		return;
	}

	/* New entry */
	struct neighbour_entry new_entry = {};
	new_entry.first_seen = now;
	new_entry.last_seen = now;
	bpf_map_update_elem(&neighbours, &key, &new_entry, BPF_NOEXIST);
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

	return TC_ACT_UNSPEC;
}
