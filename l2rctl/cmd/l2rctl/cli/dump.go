package cli

import (
	"github.com/msune/l2radar/l2rctl/internal/dump"
	"github.com/spf13/cobra"
)

var dumpOutput string

var dumpCmd = &cobra.Command{
	Use:   "dump <interface>",
	Short: "Dump neighbour table",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r := NewRunner()
		return dump.Dump(r, dump.Opts{
			Iface:  args[0],
			Output: dumpOutput,
		})
	},
}

func init() {
	dumpCmd.Flags().StringVarP(&dumpOutput, "output", "o", "", "output format (json)")

	rootCmd.AddCommand(dumpCmd)
}
