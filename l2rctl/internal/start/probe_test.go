package start

import (
	"fmt"
	"strings"
	"testing"

	"github.com/msune/l2rctl/internal/docker"
)

func TestStartProbeDefaultArgs(t *testing.T) {
	m := &docker.MockRunner{}
	opts := ProbeOpts{
		Ifaces:         []string{"any"},
		ExportDir:      "/tmp/l2radar",
		ExportInterval: "5s",
		PinPath:        "/sys/fs/bpf/l2radar",
		Image:          "ghcr.io/msune/l2radar:latest",
	}

	err := StartProbe(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Calls) < 1 {
		t.Fatal("expected at least one docker call")
	}

	// Find the "run" call
	var runCall []string
	for _, c := range m.Calls {
		if len(c) > 0 && c[0] == "run" {
			runCall = c
			break
		}
	}
	if runCall == nil {
		t.Fatal("no 'run' call found")
	}

	args := strings.Join(runCall, " ")
	for _, want := range []string{
		"--privileged",
		"--network=host",
		"-v /sys/fs/bpf:/sys/fs/bpf",
		"-v /tmp/l2radar:/tmp/l2radar",
		"--name l2radar",
		"ghcr.io/msune/l2radar:latest",
		"--iface any",
		"--export-dir /tmp/l2radar",
		"--export-interval 5s",
		"--pin-path /sys/fs/bpf/l2radar",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("missing %q in args: %s", want, args)
		}
	}
}

func TestStartProbeMultipleIfaces(t *testing.T) {
	m := &docker.MockRunner{}
	opts := ProbeOpts{
		Ifaces:         []string{"eth0", "eth1"},
		ExportDir:      "/tmp/l2radar",
		ExportInterval: "5s",
		PinPath:        "/sys/fs/bpf/l2radar",
		Image:          "ghcr.io/msune/l2radar:latest",
	}

	err := StartProbe(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var runCall []string
	for _, c := range m.Calls {
		if len(c) > 0 && c[0] == "run" {
			runCall = c
			break
		}
	}
	args := strings.Join(runCall, " ")
	if !strings.Contains(args, "--iface eth0 --iface eth1") {
		t.Errorf("missing multiple --iface flags in: %s", args)
	}
}

func TestStartProbeExtraDockerArgs(t *testing.T) {
	m := &docker.MockRunner{}
	opts := ProbeOpts{
		Ifaces:         []string{"any"},
		ExportDir:      "/tmp/l2radar",
		ExportInterval: "5s",
		PinPath:        "/sys/fs/bpf/l2radar",
		Image:          "ghcr.io/msune/l2radar:latest",
		ExtraArgs:      "--cpus 2 --memory 512m",
	}

	err := StartProbe(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var runCall []string
	for _, c := range m.Calls {
		if len(c) > 0 && c[0] == "run" {
			runCall = c
			break
		}
	}
	args := strings.Join(runCall, " ")
	if !strings.Contains(args, "--cpus 2 --memory 512m") {
		t.Errorf("missing extra args in: %s", args)
	}
}

func TestStartProbeSkipIfRunning(t *testing.T) {
	m := &docker.MockRunner{
		StdoutFn: func(args []string) string {
			if len(args) >= 2 && args[0] == "inspect" {
				return `[{"State":{"Status":"running"}}]`
			}
			return ""
		},
	}
	opts := ProbeOpts{
		Ifaces:         []string{"any"},
		ExportDir:      "/tmp/l2radar",
		ExportInterval: "5s",
		PinPath:        "/sys/fs/bpf/l2radar",
		Image:          "ghcr.io/msune/l2radar:latest",
	}

	err := StartProbe(m, opts)
	if err == nil {
		t.Fatal("expected error when container is running")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStartProbeRemoveIfStopped(t *testing.T) {
	m := &docker.MockRunner{
		StdoutFn: func(args []string) string {
			if len(args) >= 2 && args[0] == "inspect" {
				return `[{"State":{"Status":"exited"}}]`
			}
			return ""
		},
	}
	opts := ProbeOpts{
		Ifaces:         []string{"any"},
		ExportDir:      "/tmp/l2radar",
		ExportInterval: "5s",
		PinPath:        "/sys/fs/bpf/l2radar",
		Image:          "ghcr.io/msune/l2radar:latest",
	}

	err := StartProbe(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify rm was called before run
	var gotRm, gotRun bool
	for _, c := range m.Calls {
		if len(c) >= 2 && c[0] == "rm" && c[1] == "l2radar" {
			gotRm = true
		}
		if len(c) >= 1 && c[0] == "run" {
			if !gotRm {
				t.Error("'run' called before 'rm'")
			}
			gotRun = true
		}
	}
	if !gotRm {
		t.Error("missing 'rm' call")
	}
	if !gotRun {
		t.Error("missing 'run' call")
	}
}

func TestStartProbeNotFound(t *testing.T) {
	m := &docker.MockRunner{
		ErrFn: func(args []string) error {
			if len(args) >= 1 && args[0] == "inspect" {
				return fmt.Errorf("Error: No such container: l2radar")
			}
			return nil
		},
	}
	opts := ProbeOpts{
		Ifaces:         []string{"any"},
		ExportDir:      "/tmp/l2radar",
		ExportInterval: "5s",
		PinPath:        "/sys/fs/bpf/l2radar",
		Image:          "ghcr.io/msune/l2radar:latest",
	}

	err := StartProbe(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStartProbePullsBeforeRun(t *testing.T) {
	m := &docker.MockRunner{}
	opts := ProbeOpts{
		Ifaces:         []string{"any"},
		ExportDir:      "/tmp/l2radar",
		ExportInterval: "5s",
		PinPath:        "/sys/fs/bpf/l2radar",
		Image:          "ghcr.io/msune/l2radar:latest",
	}

	err := StartProbe(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var pullIdx, runIdx int
	for i, c := range m.Calls {
		if len(c) >= 1 && c[0] == "pull" {
			pullIdx = i
			args := strings.Join(c, " ")
			if !strings.Contains(args, "--quiet") {
				t.Errorf("pull missing --quiet: %s", args)
			}
			if !strings.Contains(args, "ghcr.io/msune/l2radar:latest") {
				t.Errorf("pull missing image: %s", args)
			}
		}
		if len(c) >= 1 && c[0] == "run" {
			runIdx = i
		}
	}
	if pullIdx == 0 && runIdx == 0 {
		t.Fatal("no pull or run calls found")
	}
	if pullIdx >= runIdx {
		t.Error("pull must happen before run")
	}
}

func TestStartProbePullFailure(t *testing.T) {
	m := &docker.MockRunner{
		StderrFn: func(args []string) string {
			if len(args) >= 1 && args[0] == "pull" {
				return "Error: image not found"
			}
			return ""
		},
		ErrFn: func(args []string) error {
			if len(args) >= 1 && args[0] == "pull" {
				return fmt.Errorf("exit status 1")
			}
			return nil
		},
	}
	opts := ProbeOpts{
		Ifaces:         []string{"any"},
		ExportDir:      "/tmp/l2radar",
		ExportInterval: "5s",
		PinPath:        "/sys/fs/bpf/l2radar",
		Image:          "ghcr.io/msune/l2radar:latest",
	}

	err := StartProbe(m, opts)
	if err == nil {
		t.Fatal("expected error on pull failure")
	}
	if !strings.Contains(err.Error(), "pull image") {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify run was NOT called
	for _, c := range m.Calls {
		if len(c) >= 1 && c[0] == "run" {
			t.Error("run should not be called after pull failure")
		}
	}
}

func TestStartProbeInspectUsesTypeContainer(t *testing.T) {
	m := &docker.MockRunner{
		ErrFn: func(args []string) error {
			if len(args) >= 1 && args[0] == "inspect" {
				return fmt.Errorf("Error: No such container: l2radar")
			}
			return nil
		},
	}
	opts := ProbeOpts{
		Ifaces:         []string{"any"},
		ExportDir:      "/tmp/l2radar",
		ExportInterval: "5s",
		PinPath:        "/sys/fs/bpf/l2radar",
		Image:          "ghcr.io/msune/l2radar:latest",
	}

	_ = StartProbe(m, opts)

	var inspectCall []string
	for _, c := range m.Calls {
		if len(c) > 0 && c[0] == "inspect" {
			inspectCall = c
			break
		}
	}
	if inspectCall == nil {
		t.Fatal("no 'inspect' call found")
	}
	args := strings.Join(inspectCall, " ")
	if !strings.Contains(args, "--type container") {
		t.Errorf("inspect missing --type container flag: %s", args)
	}
}

func TestStartProbeImageOnlyNoContainer(t *testing.T) {
	// Simulate: image named "l2radar" exists but no container.
	// docker inspect --type container returns an error in this case.
	m := &docker.MockRunner{
		ErrFn: func(args []string) error {
			if len(args) >= 1 && args[0] == "inspect" {
				return fmt.Errorf("Error: No such container: l2radar")
			}
			return nil
		},
	}
	opts := ProbeOpts{
		Ifaces:         []string{"any"},
		ExportDir:      "/tmp/l2radar",
		ExportInterval: "5s",
		PinPath:        "/sys/fs/bpf/l2radar",
		Image:          "ghcr.io/msune/l2radar:latest",
	}

	err := StartProbe(m, opts)
	if err != nil {
		t.Fatalf("image-only should not block start: %v", err)
	}

	// Verify no rm was called (there's no container to remove)
	for _, c := range m.Calls {
		if len(c) >= 1 && c[0] == "rm" {
			t.Error("unexpected 'rm' call when only image exists")
		}
	}
}
