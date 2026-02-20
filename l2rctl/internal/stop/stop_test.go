package stop

import (
	"fmt"
	"strings"
	"testing"

	"github.com/msune/l2radar/l2rctl/internal/docker"
)

func TestStopProbe(t *testing.T) {
	m := &docker.MockRunner{}
	err := Stop(m, Opts{Target: "probe", VolumeName: "l2radar-data"})
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

func TestStopProbeDoesNotRemoveVolume(t *testing.T) {
	m := &docker.MockRunner{}
	err := Stop(m, Opts{Target: "probe", VolumeName: "l2radar-data"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range m.Calls {
		args := strings.Join(c, " ")
		if strings.Contains(args, "volume rm") {
			t.Errorf("stop probe must not remove volume, got: %s", args)
		}
	}
}

func TestStopUI(t *testing.T) {
	m := &docker.MockRunner{}
	err := Stop(m, Opts{Target: "ui", VolumeName: "l2radar-data"})
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

func TestStopUIDoesNotRemoveVolume(t *testing.T) {
	m := &docker.MockRunner{}
	err := Stop(m, Opts{Target: "ui", VolumeName: "l2radar-data"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range m.Calls {
		args := strings.Join(c, " ")
		if strings.Contains(args, "volume rm") {
			t.Errorf("stop ui must not remove volume, got: %s", args)
		}
	}
}

func TestStopAll(t *testing.T) {
	m := &docker.MockRunner{}
	err := Stop(m, Opts{Target: "all", VolumeName: "l2radar-data"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have stop+rm for both containers, plus volume rm
	if len(m.Calls) < 5 {
		t.Errorf("expected at least 5 calls (stop+rm probe, stop+rm ui, volume rm), got %d", len(m.Calls))
	}
}

func TestStopAllRemovesVolume(t *testing.T) {
	m := &docker.MockRunner{}
	err := Stop(m, Opts{Target: "all", VolumeName: "l2radar-data"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotVolumeRm bool
	for _, c := range m.Calls {
		args := strings.Join(c, " ")
		if strings.Contains(args, "volume rm l2radar-data") {
			gotVolumeRm = true
		}
	}
	if !gotVolumeRm {
		t.Error("missing 'volume rm l2radar-data' call")
	}
}

func TestStopAllVolumeRmAfterContainers(t *testing.T) {
	m := &docker.MockRunner{}
	err := Stop(m, Opts{Target: "all", VolumeName: "l2radar-data"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Use slice matching to avoid "rm l2radar" being a substring of "volume rm l2radar-data".
	var probeRmIdx, uiRmIdx, volumeRmIdx int
	for i, c := range m.Calls {
		if len(c) == 2 && c[0] == "rm" && c[1] == ProbeContainer {
			probeRmIdx = i + 1
		}
		if len(c) == 2 && c[0] == "rm" && c[1] == UIContainer {
			uiRmIdx = i + 1
		}
		if len(c) == 3 && c[0] == "volume" && c[1] == "rm" {
			volumeRmIdx = i + 1
		}
	}
	if volumeRmIdx == 0 {
		t.Fatal("no 'volume rm' call found")
	}
	if volumeRmIdx <= probeRmIdx {
		t.Error("volume rm must happen after probe rm")
	}
	if volumeRmIdx <= uiRmIdx {
		t.Error("volume rm must happen after ui rm")
	}
}

func TestStopAllVolumeNotFound(t *testing.T) {
	m := &docker.MockRunner{
		ErrFn: func(args []string) error {
			if len(args) >= 2 && args[0] == "volume" && args[1] == "rm" {
				return fmt.Errorf("Error: No such volume: l2radar-data")
			}
			return nil
		},
	}
	err := Stop(m, Opts{Target: "all", VolumeName: "l2radar-data"})
	if err != nil {
		t.Fatalf("volume not found should be ignored, got: %v", err)
	}
}

func TestStopNotFound(t *testing.T) {
	m := &docker.MockRunner{
		ErrFn: func(args []string) error {
			return fmt.Errorf("Error: No such container: l2radar")
		},
	}
	err := Stop(m, Opts{Target: "probe", VolumeName: "l2radar-data"})
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
