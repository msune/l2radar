package dump

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/msune/l2radar/l2rctl/internal/docker"
)

const ProbeContainer = "l2radar"

// Opts holds dump command options.
type Opts struct {
	Iface     string
	Output    string
	ExportDir string
}

// Dump executes the dump command.
func Dump(r docker.Runner, opts Opts) error {
	if opts.Iface == "" {
		return fmt.Errorf("--iface is required")
	}

	if opts.Output == "json" {
		return dumpJSON(opts)
	}
	return dumpTable(r, opts)
}

func dumpTable(r docker.Runner, opts Opts) error {
	return r.RunAttached("exec", ProbeContainer, "l2radar", "dump", "--iface", opts.Iface)
}

func dumpJSON(opts Opts) error {
	path := filepath.Join(opts.ExportDir, fmt.Sprintf("neigh-%s.json", opts.Iface))
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	fmt.Print(string(data))
	return nil
}
