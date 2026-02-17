package cli

import (
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [component]",
	Short: "Install and start l2radar containers with automatic restart",
	Long: `Install and start l2radar containers with automatic restart on reboot.

Behaves like "start" but configures Docker's restart policy
(unless-stopped) so containers are automatically restarted by Docker
after a system reboot. Requires the Docker daemon to be enabled at boot
(e.g. "systemctl enable docker").

Use "l2rctl stop" to stop containers. Stopped containers will NOT be
restarted on reboot (unless-stopped policy). To fully remove the
restart policy, stop and remove the containers.

Components: all (default), probe, ui`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStartOrInstall(cmd, args, "unless-stopped")
	},
}

func init() {
	addStartFlags(installCmd)
	rootCmd.AddCommand(installCmd)
}
