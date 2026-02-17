package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/msune/l2radar/l2rctl/internal/docker"
	"github.com/msune/l2radar/l2rctl/internal/version"
	"github.com/spf13/cobra"
)

// NewRunner is overridable in tests.
var NewRunner = func() docker.Runner { return &docker.RealRunner{} }

// versionHint receives the background version-check result.
var versionHint chan string

var rootCmd = &cobra.Command{
	Use:   "l2rctl",
	Short: "Manage l2radar containers",
	Long: `l2rctl manages l2radar containers (probe and UI).

Use "l2rctl <command> --help" for more information about a command.`,
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip background check for the version subcommand (it does its own).
		if cmd.Name() == "version" {
			return
		}
		versionHint = make(chan string, 1)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			msg := version.CheckCached(ctx, version.Current(),
				version.DefaultCacheFile(), version.ProxyURL)
			versionHint <- msg
		}()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if versionHint == nil {
			return
		}
		msg := <-versionHint
		if msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
