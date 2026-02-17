package dump

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/msune/l2radar/l2rctl/internal/docker"
)

func TestDumpTableMode(t *testing.T) {
	m := &docker.MockRunner{}
	opts := Opts{
		Iface:     "eth0",
		Output:    "",
		ExportDir: "/tmp/l2radar",
	}

	err := Dump(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(m.Calls))
	}

	args := strings.Join(m.Calls[0], " ")
	for _, want := range []string{"exec", "l2radar", "l2radar", "dump", "--iface", "eth0"} {
		if !strings.Contains(args, want) {
			t.Errorf("missing %q in args: %s", want, args)
		}
	}
}

func TestDumpJSONMode(t *testing.T) {
	// Create temp file to simulate export
	dir := t.TempDir()
	content := `{"neighbours":[{"mac":"aa:bb:cc:dd:ee:ff"}]}`
	err := os.WriteFile(filepath.Join(dir, "neigh-eth0.json"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	m := &docker.MockRunner{}
	opts := Opts{
		Iface:     "eth0",
		Output:    "json",
		ExportDir: dir,
	}

	err = Dump(m, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// In JSON mode, no docker calls should be made
	if len(m.Calls) != 0 {
		t.Errorf("expected 0 docker calls in JSON mode, got %d", len(m.Calls))
	}
}

func TestDumpJSONFileMissing(t *testing.T) {
	m := &docker.MockRunner{}
	opts := Opts{
		Iface:     "eth0",
		Output:    "json",
		ExportDir: "/nonexistent/path",
	}

	err := Dump(m, opts)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDumpRequiresIface(t *testing.T) {
	m := &docker.MockRunner{}
	opts := Opts{
		Iface:     "",
		ExportDir: "/tmp/l2radar",
	}

	err := Dump(m, opts)
	if err == nil {
		t.Fatal("expected error for missing iface")
	}
}
