package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/marc/l2radar/probe/pkg/dump"
	"github.com/marc/l2radar/probe/pkg/export"
	"github.com/marc/l2radar/probe/pkg/loader"
	"github.com/spf13/cobra"
)

var (
	dumpIface   string
	dumpPinPath string
	dumpOutput  string
)

func marshalDumpJSON(iface string, ts time.Time, neighbours []dump.Neighbour) ([]byte, error) {
	ifInfo, err := export.LookupInterfaceInfo(iface)
	if err != nil {
		return nil, fmt.Errorf("lookup interface info: %w", err)
	}

	ifStats, err := export.LookupInterfaceStats(iface)
	if err != nil {
		return nil, fmt.Errorf("lookup interface stats: %w", err)
	}

	data := export.NewInterfaceData(iface, ts, 0, neighbours, ifInfo, ifStats)
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal JSON: %w", err)
	}
	return append(b, '\n'), nil
}

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump neighbour table for an interface",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		mapPath := dump.PinPath(dumpPinPath, dumpIface)
		neighbours, err := dump.ReadMap(mapPath)
		if err != nil {
			return fmt.Errorf("read map: %w", err)
		}

		dump.SortByLastSeen(neighbours)

		switch dumpOutput {
		case "table":
			dump.FormatTable(cmd.OutOrStdout(), neighbours)
		case "json":
			b, err := marshalDumpJSON(dumpIface, time.Now(), neighbours)
			if err != nil {
				return err
			}
			if _, err := cmd.OutOrStdout().Write(b); err != nil {
				return fmt.Errorf("write output: %w", err)
			}
		default:
			return fmt.Errorf("invalid output format %q (supported: table, json)", dumpOutput)
		}
		return nil
	},
}

func init() {
	dumpCmd.Flags().StringVar(&dumpIface, "iface", "", "network interface to dump (required)")
	dumpCmd.Flags().StringVar(&dumpPinPath, "pin-path", loader.DefaultPinPath, "base path for pinned eBPF maps")
	dumpCmd.Flags().StringVarP(&dumpOutput, "output", "o", "table", "output format (table|json)")
	dumpCmd.MarkFlagRequired("iface")

	rootCmd.AddCommand(dumpCmd)
}
