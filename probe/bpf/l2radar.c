// SPDX-License-Identifier: GPL-2.0
// l2radar - passive L2 neighbour monitor via TC/TCX

#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <bpf/bpf_helpers.h>

char LICENSE[] SEC("license") = "GPL";

SEC("tc")
int l2radar(struct __sk_buff *skb)
{
	return TC_ACT_UNSPEC;
}
