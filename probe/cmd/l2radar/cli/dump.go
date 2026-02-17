package cli

import (
	"fmt"
	"os"

	"github.com/marc/l2radar/probe/pkg/dump"
	"github.com/marc/l2radar/probe/pkg/loader"
	"github.com/spf13/cobra"
)

var (
	dumpIface   string
	dumpPinPath string
)

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
		dump.FormatTable(os.Stdout, neighbours)
		return nil
	},
}

func init() {
	dumpCmd.Flags().StringVar(&dumpIface, "iface", "", "network interface to dump (required)")
	dumpCmd.Flags().StringVar(&dumpPinPath, "pin-path", loader.DefaultPinPath, "base path for pinned eBPF maps")
	dumpCmd.MarkFlagRequired("iface")

	rootCmd.AddCommand(dumpCmd)
}
