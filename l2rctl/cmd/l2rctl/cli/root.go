package cli

import (
	"github.com/msune/l2radar/l2rctl/internal/docker"
	"github.com/spf13/cobra"
)

// NewRunner is overridable in tests.
var NewRunner = func() docker.Runner { return &docker.RealRunner{} }

var rootCmd = &cobra.Command{
	Use:   "l2rctl",
	Short: "Manage l2radar containers",
	Long: `l2rctl manages l2radar containers (probe and UI).

Use "l2rctl <command> --help" for more information about a command.`,
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
