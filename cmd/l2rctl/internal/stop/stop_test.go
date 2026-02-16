package stop

import (
	"fmt"
	"strings"
	"testing"

	"github.com/msune/l2rctl/internal/docker"
)

func TestStopProbe(t *testing.T) {
	m := &docker.MockRunner{}
	err := Stop(m, "probe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotStop, gotRm bool
	for _, c := range m.Calls {
		args := strings.Join(c, " ")
		if strings.Contains(args, "stop l2radar") && !strings.Contains(args, "l2radar-ui") {
			gotStop = true
		}
		if strings.Contains(args, "rm l2radar") && !strings.Contains(args, "l2radar-ui") {
			gotRm = true
		}
	}
	if !gotStop {
		t.Error("missing 'stop l2radar' call")
	}
	if !gotRm {
		t.Error("missing 'rm l2radar' call")
	}
}

func TestStopUI(t *testing.T) {
	m := &docker.MockRunner{}
	err := Stop(m, "ui")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotStop, gotRm bool
	for _, c := range m.Calls {
		args := strings.Join(c, " ")
		if strings.Contains(args, "stop l2radar-ui") {
			gotStop = true
		}
		if strings.Contains(args, "rm l2radar-ui") {
			gotRm = true
		}
	}
	if !gotStop {
		t.Error("missing 'stop l2radar-ui' call")
	}
	if !gotRm {
		t.Error("missing 'rm l2radar-ui' call")
	}
}

func TestStopAll(t *testing.T) {
	m := &docker.MockRunner{}
	err := Stop(m, "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have stop+rm for both containers
	if len(m.Calls) < 4 {
		t.Errorf("expected at least 4 calls, got %d", len(m.Calls))
	}
}

func TestStopNotFound(t *testing.T) {
	m := &docker.MockRunner{
		ErrFn: func(args []string) error {
			return fmt.Errorf("Error: No such container: l2radar")
		},
	}
	err := Stop(m, "probe")
	if err != nil {
		t.Fatalf("expected graceful handling, got: %v", err)
	}
}

func TestStopDefaultTarget(t *testing.T) {
	target, err := ParseTarget(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target != "all" {
		t.Errorf("got %q, want %q", target, "all")
	}
}
