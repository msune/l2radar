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

SEC("tc")
int l2radar(struct __sk_buff *skb)
{
	return TC_ACT_UNSPEC;
}
