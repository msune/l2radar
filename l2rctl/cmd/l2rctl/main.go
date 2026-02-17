package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/msune/l2rctl/internal/auth"
	"github.com/msune/l2rctl/internal/docker"
	"github.com/msune/l2rctl/internal/dump"
	"github.com/msune/l2rctl/internal/start"
	"github.com/msune/l2rctl/internal/status"
	"github.com/msune/l2rctl/internal/stop"
)

const usage = `Usage: l2rctl <command> [options]

Commands:
  start <component> [options]   Start l2radar containers
  stop  <component>             Stop l2radar containers
  status                        Show container status
  dump  [options]               Dump neighbour table

Components: all (default), probe, ui

Use "l2rctl <command> --help" for more information about a command.`

var validSubcommands = map[string]bool{
	"start":  true,
	"stop":   true,
	"status": true,
	"dump":   true,
}

func parseSubcommand(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("no subcommand specified")
	}
	sub := args[1]
	if sub == "--help" || sub == "-h" {
		return "help", nil
	}
	if !validSubcommands[sub] {
		return "", fmt.Errorf("unknown subcommand: %s", sub)
	}
	return sub, nil
}

// multiString implements flag.Value for repeatable string flags.
type multiString []string

func (m *multiString) String() string { return fmt.Sprintf("%v", *m) }
func (m *multiString) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func runStart(args []string, r docker.Runner) error {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: l2rctl start <component> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Components: all (default), probe, ui\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	// Probe flags
	var ifaces multiString
	fs.Var(&ifaces, "iface", "Interface to monitor (repeatable; \"any\"=external, \"all\"=all non-loopback)")
	exportDir := fs.String("export-dir", "/tmp/l2radar", "Export directory")
	exportInterval := fs.String("export-interval", "5s", "Export interval")
	pinPath := fs.String("pin-path", "/sys/fs/bpf/l2radar", "BPF pin path")
	probeImage := fs.String("probe-image", "ghcr.io/msune/l2radar:latest", "Probe image")
	probeDockerArgs := fs.String("probe-docker-args", "", "Extra docker args for probe")

	// UI flags
	tlsDir := fs.String("tls-dir", "", "TLS cert directory")
	userFile := fs.String("user-file", "", "Auth file path")
	var users multiString
	fs.Var(&users, "user", "User in user:pass format (repeatable)")
	enableHTTP := fs.Bool("enable-http", false, "Enable HTTP port 80")
	httpsPort := fs.Int("https-port", 12443, "Host port for HTTPS (mapped to container 443)")
	httpPort := fs.Int("http-port", 12080, "Host port for HTTP (mapped to container 80)")
	bind := fs.String("bind", "127.0.0.1", "Bind address for exposed ports (e.g. 0.0.0.0)")
	uiImage := fs.String("ui-image", "ghcr.io/msune/l2radar-ui:latest", "UI image")
	uiDockerArgs := fs.String("ui-docker-args", "", "Extra docker args for UI")

	if err := fs.Parse(args); err != nil {
		return err
	}

	target, err := start.ParseTarget(fs.Args())
	if err != nil {
		return err
	}

	if len(ifaces) == 0 {
		ifaces = []string{"any"}
	}

	// Auth validation
	if err := auth.ValidateFlags(*userFile, users); err != nil {
		return err
	}

	// Resolve auth file; generate random credentials if none specified
	// and the target includes the UI.
	authFile := *userFile
	var generatedCred string
	needsUI := target == "ui" || target == "all"
	if needsUI && *userFile == "" && len(users) == 0 {
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
		ExportDir:      *exportDir,
		ExportInterval: *exportInterval,
		PinPath:        *pinPath,
		Image:          *probeImage,
		ExtraArgs:      *probeDockerArgs,
	}

	uiOpts := start.UIOpts{
		ExportDir:  *exportDir,
		TLSDir:     *tlsDir,
		UserFile:   authFile,
		EnableHTTP: *enableHTTP,
		HTTPSPort:  *httpsPort,
		HTTPPort:   *httpPort,
		Bind:       *bind,
		Image:      *uiImage,
		ExtraArgs:  *uiDockerArgs,
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
		var urlUser, urlPass string
		if len(users) > 0 {
			urlUser, urlPass, _ = auth.ParseUser(users[0])
		}
		urls := start.BuildAccessURLs(*httpsPort, *httpPort, *bind, *enableHTTP, urlUser, urlPass)
		fmt.Println()
		fmt.Println("UI available at:")
		for _, u := range urls {
			fmt.Printf("  %s\n", u)
		}
	}

	return nil
}

func runStop(args []string, r docker.Runner) error {
	fs := flag.NewFlagSet("stop", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: l2rctl stop <component>\n\n")
		fmt.Fprintf(os.Stderr, "Components: all (default), probe, ui\n")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	target, err := stop.ParseTarget(fs.Args())
	if err != nil {
		return err
	}
	return stop.Stop(r, target)
}

func runStatus(r docker.Runner) error {
	out, err := status.Status(r)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

func runDump(args []string, r docker.Runner) error {
	fs := flag.NewFlagSet("dump", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: l2rctl dump [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}
	iface := fs.String("iface", "", "Interface name (required)")
	output := fs.String("o", "", "Output format (json)")
	exportDir := fs.String("export-dir", "/tmp/l2radar", "Export directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return dump.Dump(r, dump.Opts{
		Iface:     *iface,
		Output:    *output,
		ExportDir: *exportDir,
	})
}

func run(args []string) error {
	sub, err := parseSubcommand(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usage)
		return err
	}
	if sub == "help" {
		fmt.Println(usage)
		return nil
	}

	r := &docker.RealRunner{}
	subArgs := args[2:] // skip binary name and subcommand

	switch sub {
	case "start":
		return runStart(subArgs, r)
	case "stop":
		return runStop(subArgs, r)
	case "status":
		return runStatus(r)
	case "dump":
		return runDump(subArgs, r)
	}
	return nil
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
