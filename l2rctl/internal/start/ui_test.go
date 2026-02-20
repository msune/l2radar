package start

import (
	"fmt"
	"strings"
	"testing"

	"github.com/msune/l2radar/l2rctl/internal/docker"
)

func TestStartUIDefaultArgs(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
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
		"-v l2radar-data:/var/lib/l2radar:ro",
		"-p 127.0.0.1:12443:443",
		"--name l2radar-ui",
		"ghcr.io/msune/l2radar-ui:latest",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("missing %q in args: %s", want, args)
		}
	}
}

func TestStartUINoBindMount(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
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

	// The host-side must be the volume name, not a filesystem path.
	args := strings.Join(runCall, " ")
	if strings.Contains(args, "-v /var/lib/l2radar:/var/lib/l2radar") {
		t.Errorf("must not use bind mount, got: %s", args)
	}
}

func TestStartUIWithTLS(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		TLSDir:     "/etc/mycerts",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
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
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		UserFile:   "/path/to/auth.yaml",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
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
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		HTTPPort:   12080,
		Bind:       "127.0.0.1",
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
	if !strings.Contains(args, "-p 127.0.0.1:12080:80") {
		t.Errorf("missing port 12080 in: %s", args)
	}
	if !strings.Contains(args, "--enable-http") {
		t.Errorf("missing --enable-http in: %s", args)
	}
}

func TestStartUIExtraDockerArgs(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		ExtraArgs:  "--cpus 1",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
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

func TestStartUICustomPorts(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  8443,
		HTTPPort:   8080,
		Bind:       "127.0.0.1",
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
	if !strings.Contains(args, "-p 127.0.0.1:8443:443") {
		t.Errorf("missing custom HTTPS port in: %s", args)
	}
	if !strings.Contains(args, "-p 127.0.0.1:8080:80") {
		t.Errorf("missing custom HTTP port in: %s", args)
	}
}

func TestStartUIBindAllInterfaces(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		HTTPPort:   12080,
		Bind:       "0.0.0.0",
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
	if !strings.Contains(args, "-p 0.0.0.0:12443:443") {
		t.Errorf("missing 0.0.0.0 HTTPS binding in: %s", args)
	}
	if !strings.Contains(args, "-p 0.0.0.0:12080:80") {
		t.Errorf("missing 0.0.0.0 HTTP binding in: %s", args)
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
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
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
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
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
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
	}

	err := StartUI(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStartUIPullsBeforeRun(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
	}

	err := StartUI(m, opts)
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
			if !strings.Contains(args, "ghcr.io/msune/l2radar-ui:latest") {
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

func TestStartUIPullFailure(t *testing.T) {
	m := &docker.MockRunner{
		StderrFn: func(args []string) string {
			if len(args) >= 1 && args[0] == "pull" {
				return "Error: image not found"
			}
			return ""
		},
		ErrFn: func(args []string) error {
			if len(args) >= 1 && (args[0] == "pull" || args[0] == "image") {
				return fmt.Errorf("exit status 1")
			}
			return nil
		},
	}
	opts := UIOpts{
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
	}

	err := StartUI(m, opts)
	if err == nil {
		t.Fatal("expected error on pull failure")
	}
	if !strings.Contains(err.Error(), "pull image") {
		t.Errorf("unexpected error: %v", err)
	}

	for _, c := range m.Calls {
		if len(c) >= 1 && c[0] == "run" {
			t.Error("run should not be called after pull failure")
		}
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
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
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

func TestStartUIRestartPolicy(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir:     "/var/lib/l2radar",
		VolumeName:    "l2radar-data",
		Image:         "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:     12443,
		Bind:          "127.0.0.1",
		RestartPolicy: "unless-stopped",
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
	if !strings.Contains(args, "--restart unless-stopped") {
		t.Errorf("missing --restart unless-stopped in: %s", args)
	}
}

func TestStartUINoRestartPolicyByDefault(t *testing.T) {
	m := &docker.MockRunner{}
	opts := UIOpts{
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
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
	if strings.Contains(args, "--restart") {
		t.Errorf("unexpected --restart flag in: %s", args)
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
		ExportDir:  "/var/lib/l2radar",
		VolumeName: "l2radar-data",
		Image:      "ghcr.io/msune/l2radar-ui:latest",
		HTTPSPort:  12443,
		Bind:       "127.0.0.1",
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
