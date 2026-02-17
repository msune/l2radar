package start

import (
	"fmt"

	"github.com/msune/l2rctl/internal/docker"
)

// UIOpts holds flags for the UI container.
type UIOpts struct {
	ExportDir  string
	TLSDir     string
	UserFile   string
	EnableHTTP bool
	Image      string
	ExtraArgs  string
}

// StartUI starts the l2radar-ui container.
func StartUI(r docker.Runner, opts UIOpts) error {
	if err := ensureNotRunning(r, UIContainer); err != nil {
		return err
	}
	if err := pullImage(r, opts.Image); err != nil {
		return err
	}

	args := []string{"run", "-d",
		"-v", fmt.Sprintf("%s:%s:ro", opts.ExportDir, opts.ExportDir),
		"-p", "443:443",
		"--name", UIContainer,
	}

	if opts.TLSDir != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/etc/nginx/ssl:ro", opts.TLSDir))
	}

	if opts.UserFile != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/etc/l2radar/auth.yaml:ro", opts.UserFile))
	}

	if opts.EnableHTTP {
		args = append(args, "-p", "80:80")
	}

	args = append(args, splitExtraArgs(opts.ExtraArgs)...)

	args = append(args, opts.Image)

	if opts.EnableHTTP {
		args = append(args, "--enable-http")
	}

	_, _, err := r.Run(args...)
	return err
}
