package start

import (
	"fmt"

	"github.com/msune/l2radar/l2rctl/internal/docker"
)

// UIOpts holds flags for the UI container.
type UIOpts struct {
	ExportDir     string
	VolumeName    string
	TLSDir        string
	UserFile      string
	EnableHTTP    bool
	HTTPSPort     int
	HTTPPort      int
	Bind          string
	Image         string
	ExtraArgs     string
	RestartPolicy string
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
		"-v", fmt.Sprintf("%s:%s:ro", opts.VolumeName, opts.ExportDir),
		"-p", fmt.Sprintf("%s:%d:443", opts.Bind, opts.HTTPSPort),
		"--name", UIContainer,
	}

	if opts.RestartPolicy != "" {
		args = append(args, "--restart", opts.RestartPolicy)
	}

	if opts.TLSDir != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/etc/nginx/ssl:ro", opts.TLSDir))
	}

	if opts.UserFile != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/etc/l2radar/auth.yaml:ro", opts.UserFile))
	}

	if opts.EnableHTTP {
		args = append(args, "-p", fmt.Sprintf("%s:%d:80", opts.Bind, opts.HTTPPort))
	}

	args = append(args, splitExtraArgs(opts.ExtraArgs)...)

	args = append(args, opts.Image)

	if opts.EnableHTTP {
		args = append(args, "--enable-http")
	}

	_, _, err := r.Run(args...)
	return err
}
