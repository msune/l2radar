package dump

import (
	"fmt"
	"strings"
	"testing"

	"github.com/msune/l2radar/l2rctl/internal/docker"
)

func TestDumpTableMode(t *testing.T) {
	m := &docker.MockRunner{}
	opts := Opts{
		Iface:  "eth0",
		Output: "table",
	}

	err := Dump(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(m.Calls))
	}

	want := []string{"exec", "l2radar", "/l2radar", "dump", "--iface", "eth0", "-o", "table"}
	got := m.Calls[0]
	if len(got) != len(want) {
		t.Fatalf("expected args %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("arg[%d]: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestDumpDefaultModeWithoutOutputFlag(t *testing.T) {
	m := &docker.MockRunner{}
	opts := Opts{
		Iface: "eth0",
	}

	err := Dump(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(m.Calls))
	}

	want := []string{"exec", "l2radar", "/l2radar", "dump", "--iface", "eth0"}
	got := m.Calls[0]
	if len(got) != len(want) {
		t.Fatalf("expected args %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("arg[%d]: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestDumpJSONMode(t *testing.T) {
	m := &docker.MockRunner{}
	opts := Opts{
		Iface:  "eth0",
		Output: "json",
	}

	err := Dump(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(m.Calls))
	}

	want := []string{"exec", "l2radar", "/l2radar", "dump", "--iface", "eth0", "-o", "json"}
	got := m.Calls[0]
	if len(got) != len(want) {
		t.Fatalf("expected args %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("arg[%d]: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestDumpRequiresIface(t *testing.T) {
	m := &docker.MockRunner{}
	opts := Opts{
		Iface: "",
	}

	err := Dump(m, opts)
	if err == nil {
		t.Fatal("expected error for missing iface")
	}
	if !strings.Contains(err.Error(), "interface") {
		t.Errorf("expected error about interface, got: %v", err)
	}
}

func TestDumpDockerError(t *testing.T) {
	m := &docker.MockRunner{
		ErrFn: func(args []string) error {
			return fmt.Errorf("container not running")
		},
	}
	opts := Opts{
		Iface: "eth0",
	}

	err := Dump(m, opts)
	if err == nil {
		t.Fatal("expected error from docker")
	}
	if !strings.Contains(err.Error(), "container not running") {
		t.Errorf("expected docker error, got: %v", err)
	}
}

func TestDumpInvalidOutput(t *testing.T) {
	m := &docker.MockRunner{}
	opts := Opts{
		Iface:  "eth0",
		Output: "yaml",
	}

	err := Dump(m, opts)
	if err == nil {
		t.Fatal("expected error for invalid output format")
	}
	if !strings.Contains(err.Error(), "invalid output format") {
		t.Errorf("expected invalid output format error, got: %v", err)
	}
	if len(m.Calls) != 0 {
		t.Fatalf("expected no docker calls on invalid output, got %d", len(m.Calls))
	}
}
