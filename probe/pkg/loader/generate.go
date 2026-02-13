package loader

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -Wall -g" l2radar ../../bpf/l2radar.c
