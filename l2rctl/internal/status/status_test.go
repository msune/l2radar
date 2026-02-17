package status

import (
	"fmt"
	"strings"
	"testing"

	"github.com/msune/l2radar/l2rctl/internal/docker"
)

func TestStatusBothRunning(t *testing.T) {
	m := &docker.MockRunner{
		StdoutFn: func(args []string) string {
			if len(args) >= 2 && args[0] == "inspect" {
				last := args[len(args)-1]
				if last == "l2radar" {
					return `[{"State":{"Status":"running","StartedAt":"2025-06-01T12:00:00Z"}}]`
				}
				if last == "l2radar-ui" {
					return `[{"State":{"Status":"running","StartedAt":"2025-06-01T12:01:00Z"}}]`
				}
			}
			return ""
		},
	}

	out, err := Status(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "l2radar") {
		t.Errorf("missing l2radar in output: %s", out)
	}
	if !strings.Contains(out, "l2radar-ui") {
		t.Errorf("missing l2radar-ui in output: %s", out)
	}
	if !strings.Contains(out, "running") {
		t.Errorf("missing 'running' in output: %s", out)
	}
}

func TestStatusNotFound(t *testing.T) {
	m := &docker.MockRunner{
		ErrFn: func(args []string) error {
			if len(args) >= 1 && args[0] == "inspect" {
				return fmt.Errorf("Error: No such container")
			}
			return nil
		},
	}

	out, err := Status(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("missing 'not found' in output: %s", out)
	}
}

func TestStatusMixed(t *testing.T) {
	m := &docker.MockRunner{
		StdoutFn: func(args []string) string {
			if len(args) >= 2 && args[0] == "inspect" && args[len(args)-1] == "l2radar" {
				return `[{"State":{"Status":"running","StartedAt":"2025-06-01T12:00:00Z"}}]`
			}
			return ""
		},
		ErrFn: func(args []string) error {
			if len(args) >= 2 && args[0] == "inspect" && args[len(args)-1] == "l2radar-ui" {
				return fmt.Errorf("Error: No such container")
			}
			return nil
		},
	}

	out, err := Status(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "running") {
		t.Errorf("missing 'running' in output: %s", out)
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("missing 'not found' in output: %s", out)
	}
}
