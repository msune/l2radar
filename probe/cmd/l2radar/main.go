package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/marc/l2radar/probe/pkg/dump"
	"github.com/marc/l2radar/probe/pkg/loader"
)

// stringSlice implements flag.Value for repeatable string flags.
type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

// listInterfaces is overridable for testing.
var listInterfaces = defaultListInterfaces

func defaultListInterfaces() ([]net.Interface, error) {
	return net.Interfaces()
}

// resolveInterfaces expands "any" to all L2 interfaces except loopbacks.
// Non-"any" values are returned as-is. Duplicates are removed.
func resolveInterfaces(ifaces []string) ([]string, error) {
	seen := make(map[string]bool)
	var result []string
	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}

	for _, name := range ifaces {
		if strings.EqualFold(name, "any") {
			allIfaces, err := listInterfaces()
			if err != nil {
				return nil, fmt.Errorf("listing interfaces: %w", err)
			}
			for _, iface := range allIfaces {
				if iface.Flags&net.FlagLoopback != 0 {
					continue
				}
				add(iface.Name)
			}
		} else {
			add(name)
		}
	}
	return result, nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: l2radar [flags]           — attach probes\n")
	fmt.Fprintf(os.Stderr, "       l2radar dump [flags]      — dump neighbour table for an interface\n")
	fmt.Fprintf(os.Stderr, "\nRun 'l2radar -help' or 'l2radar dump -help' for details.\n")
}

func main() {
	// Check if first arg is the "dump" subcommand.
	if len(os.Args) >= 2 && os.Args[1] == "dump" {
		runDump(os.Args[2:])
		return
	}

	// Default mode: attach probes.
	runDefault(os.Args[1:])
}

func runDefault(args []string) {
	fs := flag.NewFlagSet("l2radar", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: l2radar [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Attach eBPF probes to network interfaces.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}

	var ifaces stringSlice
	pinPath := fs.String("pin-path", loader.DefaultPinPath, "base path for pinning eBPF maps")
	fs.Var(&ifaces, "iface", "network interface to monitor (repeatable; \"any\" for all L2 interfaces)")
	fs.Parse(args)

	if len(ifaces) == 0 {
		fmt.Fprintf(os.Stderr, "error: at least one -iface is required\n")
		fs.Usage()
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Resolve "any" to actual interface names.
	resolved, err := resolveInterfaces(ifaces)
	if err != nil {
		logger.Error("failed to resolve interfaces", "error", err)
		os.Exit(1)
	}
	if len(resolved) == 0 {
		logger.Error("no interfaces found")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var probes []*loader.Probe
	for _, iface := range resolved {
		probe, err := loader.Attach(iface, *pinPath, logger)
		if err != nil {
			logger.Error("failed to attach probe", "interface", iface, "error", err)
			for _, p := range probes {
				p.Close()
			}
			os.Exit(1)
		}
		probes = append(probes, probe)
	}

	logger.Info("l2radar running", "interfaces", resolved, "pin_path", *pinPath)

	<-ctx.Done()
	logger.Info("shutting down...")

	for _, p := range probes {
		if err := p.Close(); err != nil {
			logger.Error("failed to close probe", "interface", p.Interface(), "error", err)
		}
	}
}

func runDump(args []string) {
	fs := flag.NewFlagSet("dump", flag.ExitOnError)
	iface := fs.String("iface", "", "network interface to dump")
	pinPath := fs.String("pin-path", loader.DefaultPinPath, "base path for pinned eBPF maps")
	fs.Parse(args)

	if *iface == "" {
		fmt.Fprintf(os.Stderr, "error: -iface is required\n")
		fs.Usage()
		os.Exit(1)
	}

	mapPath := dump.PinPath(*pinPath, *iface)
	neighbours, err := dump.ReadMap(mapPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	dump.SortByLastSeen(neighbours)
	dump.FormatTable(os.Stdout, neighbours)
}
