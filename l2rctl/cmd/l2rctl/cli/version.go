package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/msune/l2radar/l2rctl/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print l2rctl version and check for updates",
	RunE: func(cmd *cobra.Command, args []string) error {
		cur := version.Current()
		fmt.Fprintf(cmd.OutOrStdout(), "l2rctl %s\n", cur)

		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()

		latest, err := version.FetchLatest(ctx, version.ProxyURL)
		if err != nil {
			return nil // silently ignore network errors
		}

		if version.IsNewer(cur, latest) {
			fmt.Fprintln(cmd.ErrOrStderr(), version.UpgradeMsg(latest))
		} else if cur != "(devel)" {
			fmt.Fprintln(cmd.OutOrStdout(), "You are up to date.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
