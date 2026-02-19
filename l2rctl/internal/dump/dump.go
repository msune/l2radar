package dump

import (
	"fmt"

	"github.com/msune/l2radar/l2rctl/internal/docker"
)

const ProbeContainer = "l2radar"

// Opts holds dump command options.
type Opts struct {
	Iface  string
	Output string
}

// Dump executes the dump command.
func Dump(r docker.Runner, opts Opts) error {
	if opts.Iface == "" {
		return fmt.Errorf("interface name is required")
	}
	if opts.Output != "" && opts.Output != "table" && opts.Output != "json" {
		return fmt.Errorf("invalid output format %q (supported: table, json)", opts.Output)
	}

	args := []string{"exec", ProbeContainer, "/l2radar", "dump", "--iface", opts.Iface}
	if opts.Output != "" {
		args = append(args, "-o", opts.Output)
	}
	return r.RunAttached(args...)
}
