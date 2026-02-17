package start

import (
	"fmt"

	"github.com/msune/l2rctl/internal/docker"
)

// ProbeOpts holds flags for the probe container.
type ProbeOpts struct {
	Ifaces         []string
	ExportDir      string
	ExportInterval string
	PinPath        string
	Image          string
	ExtraArgs      string
}

// StartProbe starts the l2radar probe container.
func StartProbe(r docker.Runner, opts ProbeOpts) error {
	if err := ensureNotRunning(r, ProbeContainer); err != nil {
		return err
	}
	if err := pullImage(r, opts.Image); err != nil {
		return err
	}

	args := []string{"run", "-d",
		"--privileged",
		"--network=host",
		"-v", "/sys/fs/bpf:/sys/fs/bpf",
		"-v", fmt.Sprintf("%s:%s", opts.ExportDir, opts.ExportDir),
		"--name", ProbeContainer,
	}

	args = append(args, splitExtraArgs(opts.ExtraArgs)...)

	args = append(args, opts.Image)

	// Container command args
	for _, iface := range opts.Ifaces {
		args = append(args, "--iface", iface)
	}
	args = append(args, "--export-dir", opts.ExportDir)
	args = append(args, "--export-interval", opts.ExportInterval)
	args = append(args, "--pin-path", opts.PinPath)

	_, _, err := r.Run(args...)
	return err
}
