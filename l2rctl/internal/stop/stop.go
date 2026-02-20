package stop

import (
	"fmt"
	"strings"

	"github.com/msune/l2radar/l2rctl/internal/docker"
)

const (
	ProbeContainer = "l2radar"
	UIContainer    = "l2radar-ui"
)

var validTargets = map[string]bool{
	"all":   true,
	"probe": true,
	"ui":    true,
}

// Opts holds options for the Stop function.
type Opts struct {
	Target     string
	VolumeName string
}

// ParseTarget extracts the target from args (default: "all").
func ParseTarget(args []string) (string, error) {
	if len(args) == 0 {
		return "all", nil
	}
	t := args[0]
	if !validTargets[t] {
		return "", fmt.Errorf("invalid target: %s (must be all, probe, or ui)", t)
	}
	return t, nil
}

// isNotFound returns true if the error indicates a resource was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "No such container") ||
		strings.Contains(err.Error(), "No such volume") ||
		strings.Contains(err.Error(), "not found")
}

// removeVolume removes the named Docker volume, ignoring not-found errors.
func removeVolume(r docker.Runner, name string) error {
	if name == "" {
		return nil
	}
	_, _, err := r.Run("volume", "rm", name)
	if isNotFound(err) {
		return nil
	}
	return err
}

// stopContainer stops and removes a single container, ignoring not-found.
func stopContainer(r docker.Runner, name string) error {
	if _, _, err := r.Run("stop", name); err != nil {
		if !isNotFound(err) {
			return fmt.Errorf("stop %s: %w", name, err)
		}
	}
	if _, _, err := r.Run("rm", name); err != nil {
		if !isNotFound(err) {
			return fmt.Errorf("rm %s: %w", name, err)
		}
	}
	return nil
}

// Stop stops and removes target containers.
func Stop(r docker.Runner, opts Opts) error {
	switch opts.Target {
	case "probe":
		return stopContainer(r, ProbeContainer)
	case "ui":
		return stopContainer(r, UIContainer)
	case "all":
		if err := stopContainer(r, ProbeContainer); err != nil {
			return err
		}
		if err := stopContainer(r, UIContainer); err != nil {
			return err
		}
		return removeVolume(r, opts.VolumeName)
	default:
		return fmt.Errorf("invalid target: %s", opts.Target)
	}
}
