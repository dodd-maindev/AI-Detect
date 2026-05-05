// clang -O2 -g -target bpf -c xdp_firewall.c -o xdp_firewall.o
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/tcp.h>
#include <bpf/bpf_helpers.h>

// BPF Hash Map: Key = Source IPv4 (__u32), Value = 1 (blocked)
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 1024);
	__type(key, __u32);
	__type(value, __u32);
} blocked_ips SEC(".maps");

SEC("xdp")
int xdp_firewall(struct xdp_md *ctx) {
	void *data     = (void *)(long)ctx->data;
	void *data_end = (void *)(long)ctx->data_end;

	// Parse Ethernet header
	struct ethhdr *eth = data;
	if ((void *)(eth + 1) > data_end)
		return XDP_PASS;

	// Only process IPv4 packets
	if (eth->h_proto != __constant_htons(ETH_P_IP))
		return XDP_PASS;

	// Parse IP header
	struct iphdr *ip = (void *)(eth + 1);
	if ((void *)(ip + 1) > data_end)
		return XDP_PASS;

	// SAFETY: Always allow SSH (Port 22) to prevent self-lockout
	if (ip->protocol == IPPROTO_TCP) {
		struct tcphdr *tcp = (void *)ip + (ip->ihl * 4);
		if ((void *)(tcp + 1) <= data_end) {
			if (tcp->dest == __constant_htons(22) ||
			    tcp->source == __constant_htons(22))
				return XDP_PASS;
		}
	}

	// Lookup source IP in blocked list
	__u32 src_ip = ip->saddr;
	__u32 *blocked = bpf_map_lookup_elem(&blocked_ips, &src_ip);
	if (blocked)
		return XDP_DROP;

	return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
