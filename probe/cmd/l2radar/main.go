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
	"time"

	"github.com/marc/l2radar/probe/pkg/dump"
	"github.com/marc/l2radar/probe/pkg/export"
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

// virtualPrefixes lists interface name prefixes considered virtual/infrastructure.
// These are filtered out by the "any" keyword.
var virtualPrefixes = []string{"veth", "docker", "br-", "virbr"}

// isVirtualInterface returns true if the interface name matches a virtual prefix.
func isVirtualInterface(name string) bool {
	for _, prefix := range virtualPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// resolveInterfaces expands "any" and "all" keywords to matching interfaces.
//   - "all": all L2 interfaces except loopbacks.
//   - "any": all external interfaces (excludes loopbacks and virtual interfaces
//     like docker*, veth*, br-*, virbr*).
//
// Non-keyword values are returned as-is. Duplicates are removed.
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
		lower := strings.ToLower(name)
		if lower == "any" || lower == "all" {
			filterVirtual := lower == "any"
			allIfaces, err := listInterfaces()
			if err != nil {
				return nil, fmt.Errorf("listing interfaces: %w", err)
			}
			for _, iface := range allIfaces {
				if iface.Flags&net.FlagLoopback != 0 {
					continue
				}
				if filterVirtual && isVirtualInterface(iface.Name) {
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
	fmt.Fprintf(os.Stderr, "Usage: l2radar [flags]           — attach probes and optionally export\n")
	fmt.Fprintf(os.Stderr, "       l2radar dump [flags]      — dump neighbour table for an interface\n")
	fmt.Fprintf(os.Stderr, "\nRun 'l2radar -help' or 'l2radar dump -help' for details.\n")
}

func main() {
	// Check if first arg is the "dump" subcommand.
	if len(os.Args) >= 2 && os.Args[1] == "dump" {
		runDump(os.Args[2:])
		return
	}

	// Default mode: attach probes (+ optional export).
	runDefault(os.Args[1:])
}

func runDefault(args []string) {
	fs := flag.NewFlagSet("l2radar", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: l2radar [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Attach eBPF probes to network interfaces and optionally export JSON.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}

	var ifaces stringSlice
	pinPath := fs.String("pin-path", loader.DefaultPinPath, "base path for pinning eBPF maps")
	exportDir := fs.String("export-dir", "", "directory to write JSON files (disabled if empty)")
	exportInterval := fs.Duration("export-interval", 5*time.Second, "export interval (only used with -export-dir)")
	fs.Var(&ifaces, "iface", "network interface to monitor (repeatable; \"any\" for external interfaces, \"all\" for all L2 interfaces)")
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

	// Attach probes to all interfaces.
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

	// If export is enabled, start the export loop.
	if *exportDir != "" {
		if *exportInterval <= 0 {
			logger.Error("export-interval must be positive")
			for _, p := range probes {
				p.Close()
			}
			os.Exit(1)
		}

		if err := os.MkdirAll(*exportDir, 0755); err != nil {
			logger.Error("failed to create export directory", "dir", *exportDir, "error", err)
			for _, p := range probes {
				p.Close()
			}
			os.Exit(1)
		}

		logger.Info("export enabled", "dir", *exportDir, "interval", exportInterval.String())

		ticker := time.NewTicker(*exportInterval)
		defer ticker.Stop()

		// Export immediately, then on each tick.
		exportAll(resolved, *pinPath, *exportDir, *exportInterval, logger)
		for {
			select {
			case <-ctx.Done():
				goto shutdown
			case <-ticker.C:
				exportAll(resolved, *pinPath, *exportDir, *exportInterval, logger)
			}
		}
	} else {
		// No export — just wait for signal.
		<-ctx.Done()
	}

shutdown:
	logger.Info("shutting down...")
	for _, p := range probes {
		if err := p.Close(); err != nil {
			logger.Error("failed to close probe", "interface", p.Interface(), "error", err)
		}
	}
}

func exportAll(ifaces []string, pinPath, outputDir string, interval time.Duration, logger *slog.Logger) {
	for _, iface := range ifaces {
		mapPath := dump.PinPath(pinPath, iface)
		neighbours, err := dump.ReadMap(mapPath)
		if err != nil {
			logger.Error("failed to read map", "interface", iface, "error", err)
			continue
		}

		dump.SortByLastSeen(neighbours)

		ifInfo, err := export.LookupInterfaceInfo(iface)
		if err != nil {
			logger.Warn("failed to lookup interface info", "interface", iface, "error", err)
		}

		ifStats, err := export.LookupInterfaceStats(iface)
		if err != nil {
			logger.Warn("failed to lookup interface stats", "interface", iface, "error", err)
		}

		if err := export.WriteJSON(iface, neighbours, outputDir, time.Now(), interval, ifInfo, ifStats); err != nil {
			logger.Error("failed to write JSON", "interface", iface, "error", err)
			continue
		}

		logger.Debug("exported", "interface", iface, "neighbours", len(neighbours))
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
