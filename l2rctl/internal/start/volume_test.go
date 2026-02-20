package start

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/msune/l2radar/l2rctl/internal/docker"
)

// mockForVolume builds a MockRunner that reports containers as not-found
// and the volume as existing or not based on volumeExists.
func mockForVolume(volumeExists bool) *docker.MockRunner {
	return &docker.MockRunner{
		ErrFn: func(args []string) error {
			if len(args) >= 1 && args[0] == "inspect" {
				return fmt.Errorf("Error: No such container")
			}
			if !volumeExists && len(args) >= 2 && args[0] == "volume" && args[1] == "inspect" {
				return fmt.Errorf("Error: No such volume: l2radar-data")
			}
			return nil
		},
		StdoutFn: func(args []string) string {
			if volumeExists && len(args) >= 2 && args[0] == "volume" && args[1] == "inspect" {
				return `[{"Name":"l2radar-data"}]`
			}
			return ""
		},
	}
}

func TestEnsureCleanVolume_StaleVolume(t *testing.T) {
	// Neither container running; volume exists → warn + rm.
	m := mockForVolume(true)
	var warn bytes.Buffer

	if err := EnsureCleanVolume(m, "l2radar-data", &warn); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotVolumeRm bool
	for _, c := range m.Calls {
		if len(c) == 3 && c[0] == "volume" && c[1] == "rm" && c[2] == "l2radar-data" {
			gotVolumeRm = true
		}
	}
	if !gotVolumeRm {
		t.Error("expected 'volume rm l2radar-data' for stale volume")
	}
	if !strings.Contains(warn.String(), "l2radar-data") {
		t.Errorf("expected warning mentioning volume name, got: %q", warn.String())
	}
}

func TestEnsureCleanVolume_NoVolume(t *testing.T) {
	// Volume does not exist → no rm, no error.
	m := mockForVolume(false)
	var warn bytes.Buffer

	if err := EnsureCleanVolume(m, "l2radar-data", &warn); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range m.Calls {
		if len(c) >= 2 && c[0] == "volume" && c[1] == "rm" {
			t.Errorf("must not call volume rm when volume does not exist, got: %v", c)
		}
	}
}

func TestEnsureCleanVolume_ProbeRunning(t *testing.T) {
	// Probe is running → skip volume rm even if volume exists.
	m := &docker.MockRunner{
		StdoutFn: func(args []string) string {
			if len(args) >= 2 && args[0] == "inspect" && args[len(args)-1] == ProbeContainer {
				return `[{"State":{"Status":"running"}}]`
			}
			if len(args) >= 2 && args[0] == "volume" && args[1] == "inspect" {
				return `[{"Name":"l2radar-data"}]`
			}
			return ""
		},
	}
	var warn bytes.Buffer

	if err := EnsureCleanVolume(m, "l2radar-data", &warn); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range m.Calls {
		if len(c) >= 2 && c[0] == "volume" && c[1] == "rm" {
			t.Errorf("must not rm volume while probe is running, got: %v", c)
		}
	}
}

func TestEnsureCleanVolume_UIRunning(t *testing.T) {
	// UI is running → skip volume rm even if volume exists.
	m := &docker.MockRunner{
		StdoutFn: func(args []string) string {
			if len(args) >= 2 && args[0] == "inspect" && args[len(args)-1] == UIContainer {
				return `[{"State":{"Status":"running"}}]`
			}
			if len(args) >= 2 && args[0] == "volume" && args[1] == "inspect" {
				return `[{"Name":"l2radar-data"}]`
			}
			return ""
		},
	}
	var warn bytes.Buffer

	if err := EnsureCleanVolume(m, "l2radar-data", &warn); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range m.Calls {
		if len(c) >= 2 && c[0] == "volume" && c[1] == "rm" {
			t.Errorf("must not rm volume while UI is running, got: %v", c)
		}
	}
}
