package cli

import (
	"github.com/msune/l2radar/l2rctl/internal/dump"
	"github.com/spf13/cobra"
)

var (
	dumpIface     string
	dumpOutput    string
	dumpExportDir string
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump neighbour table",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r := NewRunner()
		return dump.Dump(r, dump.Opts{
			Iface:     dumpIface,
			Output:    dumpOutput,
			ExportDir: dumpExportDir,
		})
	},
}

func init() {
	dumpCmd.Flags().StringVar(&dumpIface, "iface", "", "interface name (required)")
	dumpCmd.Flags().StringVarP(&dumpOutput, "output", "o", "", "output format (json)")
	dumpCmd.Flags().StringVar(&dumpExportDir, "export-dir", "/tmp/l2radar", "export directory")
	dumpCmd.MarkFlagRequired("iface")

	rootCmd.AddCommand(dumpCmd)
}
