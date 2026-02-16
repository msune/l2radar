package start

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/msune/l2rctl/internal/docker"
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

// containerState holds docker inspect state.
type containerState struct {
	State struct {
		Status string `json:"Status"`
	} `json:"State"`
}

// checkContainer checks if a container exists and its state.
// Returns: "running", "stopped", or "notfound".
func checkContainer(r docker.Runner, name string) (string, error) {
	stdout, _, err := r.Run("inspect", name)
	if err != nil {
		return "notfound", nil
	}
	var states []containerState
	if err := json.Unmarshal([]byte(stdout), &states); err != nil {
		return "notfound", nil
	}
	if len(states) == 0 {
		return "notfound", nil
	}
	status := states[0].State.Status
	if status == "running" {
		return "running", nil
	}
	return "stopped", nil
}

// ensureNotRunning checks container state and removes stopped containers.
func ensureNotRunning(r docker.Runner, name string) error {
	state, err := checkContainer(r, name)
	if err != nil {
		return err
	}
	switch state {
	case "running":
		return fmt.Errorf("container %q is already running (stop it first)", name)
	case "stopped":
		if _, _, err := r.Run("rm", name); err != nil {
			return fmt.Errorf("remove stopped container %q: %w", name, err)
		}
	}
	return nil
}

// splitExtraArgs splits a space-separated string into args.
func splitExtraArgs(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Fields(s)
}
