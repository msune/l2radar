package cli

import (
	"github.com/msune/l2rctl/internal/stop"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [component]",
	Short: "Stop l2radar containers",
	Long:  "Stop l2radar containers.\n\nComponents: all (default), probe, ui",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r := NewRunner()
		target, err := stop.ParseTarget(args)
		if err != nil {
			return err
		}
		return stop.Stop(r, target)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
