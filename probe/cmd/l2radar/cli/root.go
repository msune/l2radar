package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marc/l2radar/probe/pkg/dump"
	"github.com/marc/l2radar/probe/pkg/export"
	"github.com/marc/l2radar/probe/pkg/loader"
	"github.com/spf13/cobra"
)

var (
	rootIfaces         []string
	rootPinPath        string
	rootExportDir      string
	rootExportInterval time.Duration
)

var rootCmd = &cobra.Command{
	Use:          "l2radar",
	Short:        "Passive L2 neighbour monitor using eBPF",
	Long:         "Attach eBPF probes to network interfaces and optionally export JSON.",
	SilenceUsage: true,
	RunE:         runRoot,
}

func init() {
	rootCmd.Flags().StringArrayVar(&rootIfaces, "iface", nil, "network interface to monitor (repeatable; \"external\" for external, \"any\" for all L2)")
	rootCmd.Flags().StringVar(&rootPinPath, "pin-path", loader.DefaultPinPath, "base path for pinning eBPF maps")
	rootCmd.Flags().StringVar(&rootExportDir, "export-dir", "", "directory to write JSON files (disabled if empty)")
	rootCmd.Flags().DurationVar(&rootExportInterval, "export-interval", 5*time.Second, "export interval (only used with --export-dir)")
	rootCmd.MarkFlagRequired("iface")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func runRoot(cmd *cobra.Command, args []string) error {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Resolve "external"/"any" to actual interface names.
	resolved, err := resolveInterfaces(rootIfaces)
	if err != nil {
		return fmt.Errorf("failed to resolve interfaces: %w", err)
	}
	if len(resolved) == 0 {
		return fmt.Errorf("no interfaces found")
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Attach probes to all interfaces.
	var probes []*loader.Probe
	for _, iface := range resolved {
		probe, err := loader.Attach(iface, rootPinPath, logger)
		if err != nil {
			for _, p := range probes {
				p.Close()
			}
			return fmt.Errorf("failed to attach probe to %s: %w", iface, err)
		}
		probes = append(probes, probe)
	}

	logger.Info("l2radar running", "interfaces", resolved, "pin_path", rootPinPath)

	// If export is enabled, start the export loop.
	if rootExportDir != "" {
		if rootExportInterval <= 0 {
			for _, p := range probes {
				p.Close()
			}
			return fmt.Errorf("export-interval must be positive")
		}

		if err := os.MkdirAll(rootExportDir, 0755); err != nil {
			for _, p := range probes {
				p.Close()
			}
			return fmt.Errorf("failed to create export directory %s: %w", rootExportDir, err)
		}

		logger.Info("export enabled", "dir", rootExportDir, "interval", rootExportInterval.String())

		ticker := time.NewTicker(rootExportInterval)
		defer ticker.Stop()

		// Export immediately, then on each tick.
		exportAll(resolved, rootPinPath, rootExportDir, rootExportInterval, logger)
		for {
			select {
			case <-ctx.Done():
				goto shutdown
			case <-ticker.C:
				exportAll(resolved, rootPinPath, rootExportDir, rootExportInterval, logger)
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
	return nil
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
