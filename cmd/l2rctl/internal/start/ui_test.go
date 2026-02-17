package start

import (
	"fmt"
	"strings"
	"testing"

	"github.com/msune/l2rctl/internal/docker"
)

func TestStartUIDefaultArgs(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir: "/tmp/l2radar",
		Image:     "ghcr.io/msune/l2radar-ui:latest",
	}

	err := StartUI(m, opts)
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
	if runCall == nil {
		t.Fatal("no 'run' call found")
	}

	args := strings.Join(runCall, " ")
	for _, want := range []string{
		"-v /tmp/l2radar:/tmp/l2radar:ro",
		"-p 443:443",
		"--name l2radar-ui",
		"ghcr.io/msune/l2radar-ui:latest",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("missing %q in args: %s", want, args)
		}
	}
}

func TestStartUIWithTLS(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir: "/tmp/l2radar",
		Image:     "ghcr.io/msune/l2radar-ui:latest",
		TLSDir:    "/etc/mycerts",
	}

	err := StartUI(m, opts)
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
	if !strings.Contains(args, "-v /etc/mycerts:/etc/nginx/ssl:ro") {
		t.Errorf("missing TLS mount in: %s", args)
	}
}

func TestStartUIWithAuthFile(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir: "/tmp/l2radar",
		Image:     "ghcr.io/msune/l2radar-ui:latest",
		UserFile:  "/path/to/auth.yaml",
	}

	err := StartUI(m, opts)
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
	if !strings.Contains(args, "-v /path/to/auth.yaml:/etc/l2radar/auth.yaml:ro") {
		t.Errorf("missing auth mount in: %s", args)
	}
}

func TestStartUIEnableHTTP(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir:  "/tmp/l2radar",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		EnableHTTP: true,
	}

	err := StartUI(m, opts)
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
	if !strings.Contains(args, "-p 80:80") {
		t.Errorf("missing port 80 in: %s", args)
	}
	if !strings.Contains(args, "--enable-http") {
		t.Errorf("missing --enable-http in: %s", args)
	}
}

func TestStartUIExtraDockerArgs(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir: "/tmp/l2radar",
		Image:     "ghcr.io/msune/l2radar-ui:latest",
		ExtraArgs: "--cpus 1",
	}

	err := StartUI(m, opts)
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
	if !strings.Contains(args, "--cpus 1") {
		t.Errorf("missing extra args in: %s", args)
	}
}

func TestStartUISkipIfRunning(t *testing.T) {
	m := &docker.MockRunner{
		StdoutFn: func(args []string) string {
			if len(args) >= 2 && args[0] == "inspect" {
				return `[{"State":{"Status":"running"}}]`
			}
			return ""
		},
	}
	opts := UIOpts{
		ExportDir: "/tmp/l2radar",
		Image:     "ghcr.io/msune/l2radar-ui:latest",
	}

	err := StartUI(m, opts)
	if err == nil {
		t.Fatal("expected error when container is running")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStartUIRemoveIfStopped(t *testing.T) {
	m := &docker.MockRunner{
		StdoutFn: func(args []string) string {
			if len(args) >= 2 && args[0] == "inspect" {
				return `[{"State":{"Status":"exited"}}]`
			}
			return ""
		},
	}
	opts := UIOpts{
		ExportDir: "/tmp/l2radar",
		Image:     "ghcr.io/msune/l2radar-ui:latest",
	}

	err := StartUI(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotRm bool
	for _, c := range m.Calls {
		if len(c) >= 2 && c[0] == "rm" && c[1] == "l2radar-ui" {
			gotRm = true
		}
	}
	if !gotRm {
		t.Error("missing 'rm' call for stopped container")
	}
}

func TestStartUINotFound(t *testing.T) {
	m := &docker.MockRunner{
		ErrFn: func(args []string) error {
			if len(args) >= 1 && args[0] == "inspect" {
				return fmt.Errorf("Error: No such container")
			}
			return nil
		},
	}
	opts := UIOpts{
		ExportDir: "/tmp/l2radar",
		Image:     "ghcr.io/msune/l2radar-ui:latest",
	}

	err := StartUI(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStartUIInspectUsesTypeContainer(t *testing.T) {
	m := &docker.MockRunner{
		ErrFn: func(args []string) error {
			if len(args) >= 1 && args[0] == "inspect" {
				return fmt.Errorf("Error: No such container")
			}
			return nil
		},
	}
	opts := UIOpts{
		ExportDir: "/tmp/l2radar",
		Image:     "ghcr.io/msune/l2radar-ui:latest",
	}

	_ = StartUI(m, opts)

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

func TestStartUIImageOnlyNoContainer(t *testing.T) {
	m := &docker.MockRunner{
		ErrFn: func(args []string) error {
			if len(args) >= 1 && args[0] == "inspect" {
				return fmt.Errorf("Error: No such container")
			}
			return nil
		},
	}
	opts := UIOpts{
		ExportDir: "/tmp/l2radar",
		Image:     "ghcr.io/msune/l2radar-ui:latest",
	}

	err := StartUI(m, opts)
	if err != nil {
		t.Fatalf("image-only should not block start: %v", err)
	}

	for _, c := range m.Calls {
		if len(c) >= 1 && c[0] == "rm" {
			t.Error("unexpected 'rm' call when only image exists")
		}
	}
}
