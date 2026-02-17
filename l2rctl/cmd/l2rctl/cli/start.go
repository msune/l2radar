package cli

import (
	"fmt"
	"os"

	"github.com/msune/l2rctl/internal/auth"
	"github.com/msune/l2rctl/internal/start"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [component]",
	Short: "Start l2radar containers",
	Long:  "Start l2radar containers.\n\nComponents: all (default), probe, ui",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runStart,
}

// Probe flags
var (
	startIfaces          []string
	startExportDir       string
	startExportInterval  string
	startPinPath         string
	startProbeImage      string
	startProbeDockerArgs string
)

// UI flags
var (
	startTLSDir       string
	startUserFile     string
	startUsers        []string
	startEnableHTTP   bool
	startHTTPSPort    int
	startHTTPPort     int
	startBind         string
	startUIImage      string
	startUIDockerArgs string
)

func init() {
	// Probe flags
	startCmd.Flags().StringArrayVar(&startIfaces, "iface", nil, "interface to monitor (repeatable; \"any\"=external, \"all\"=all non-loopback)")
	startCmd.Flags().StringVar(&startExportDir, "export-dir", "/tmp/l2radar", "export directory")
	startCmd.Flags().StringVar(&startExportInterval, "export-interval", "5s", "export interval")
	startCmd.Flags().StringVar(&startPinPath, "pin-path", "/sys/fs/bpf/l2radar", "BPF pin path")
	startCmd.Flags().StringVar(&startProbeImage, "probe-image", "ghcr.io/msune/l2radar:latest", "probe image")
	startCmd.Flags().StringVar(&startProbeDockerArgs, "probe-docker-args", "", "extra docker args for probe")

	// UI flags
	startCmd.Flags().StringVar(&startTLSDir, "tls-dir", "", "TLS cert directory")
	startCmd.Flags().StringVar(&startUserFile, "user-file", "", "auth file path")
	startCmd.Flags().StringArrayVar(&startUsers, "user", nil, "user in user:pass format (repeatable)")
	startCmd.Flags().BoolVar(&startEnableHTTP, "enable-http", false, "enable HTTP port 80")
	startCmd.Flags().IntVar(&startHTTPSPort, "https-port", 12443, "host port for HTTPS (mapped to container 443)")
	startCmd.Flags().IntVar(&startHTTPPort, "http-port", 12080, "host port for HTTP (mapped to container 80)")
	startCmd.Flags().StringVar(&startBind, "bind", "127.0.0.1", "bind address for exposed ports (e.g. 0.0.0.0)")
	startCmd.Flags().StringVar(&startUIImage, "ui-image", "ghcr.io/msune/l2radar-ui:latest", "UI image")
	startCmd.Flags().StringVar(&startUIDockerArgs, "ui-docker-args", "", "extra docker args for UI")

	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	r := NewRunner()

	target, err := start.ParseTarget(args)
	if err != nil {
		return err
	}

	ifaces := startIfaces
	if len(ifaces) == 0 {
		ifaces = []string{"any"}
	}

	// Auth validation
	if err := auth.ValidateFlags(startUserFile, startUsers); err != nil {
		return err
	}

	// Resolve auth file; generate random credentials if none specified
	// and the target includes the UI.
	authFile := startUserFile
	users := startUsers
	var generatedCred string
	needsUI := target == "ui" || target == "all"
	if needsUI && startUserFile == "" && len(users) == 0 {
		cred, err := auth.GenerateRandomCredentials()
		if err != nil {
			return err
		}
		generatedCred = cred
		users = []string{cred}
	}
	if len(users) > 0 {
		path, err := auth.WriteAuthFile(users)
		if err != nil {
			return err
		}
		defer os.Remove(path)
		authFile = path
	}

	probeOpts := start.ProbeOpts{
		Ifaces:         ifaces,
		ExportDir:      startExportDir,
		ExportInterval: startExportInterval,
		PinPath:        startPinPath,
		Image:          startProbeImage,
		ExtraArgs:      startProbeDockerArgs,
	}

	uiOpts := start.UIOpts{
		ExportDir:  startExportDir,
		TLSDir:     startTLSDir,
		UserFile:   authFile,
		EnableHTTP: startEnableHTTP,
		HTTPSPort:  startHTTPSPort,
		HTTPPort:   startHTTPPort,
		Bind:       startBind,
		Image:      startUIImage,
		ExtraArgs:  startUIDockerArgs,
	}

	switch target {
	case "probe":
		return start.StartProbe(r, probeOpts)
	case "ui":
		if err := start.StartUI(r, uiOpts); err != nil {
			return err
		}
	case "all":
		if err := start.StartProbe(r, probeOpts); err != nil {
			return err
		}
		if err := start.StartUI(r, uiOpts); err != nil {
			return err
		}
	}

	if generatedCred != "" {
		user, pass, _ := auth.ParseUser(generatedCred)
		fmt.Println()
		fmt.Println("Generated UI credentials (no --user or --user-file provided):")
		fmt.Printf("  Username: %s\n", user)
		fmt.Printf("  Password: %s\n", pass)
	}

	if needsUI {
		urls := start.BuildAccessURLs(startHTTPSPort, startHTTPPort, startBind, startEnableHTTP)
		fmt.Println()
		fmt.Println("UI available at:")
		for _, u := range urls {
			fmt.Printf("  %s\n", u)
		}
	}

	return nil
}
