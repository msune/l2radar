package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPinPathFormat(t *testing.T) {
	// Verify the expected pin path format
	iface := "eth0"
	pinBase := DefaultPinPath
	expected := filepath.Join(pinBase, fmt.Sprintf("neigh-%s", iface))
	if expected != "/sys/fs/bpf/l2radar/neigh-eth0" {
		t.Errorf("unexpected pin path: %s", expected)
	}
}

func TestPinPathMultipleInterfaces(t *testing.T) {
	ifaces := []string{"eth0", "wlan0", "br-lan"}
	pinBase := DefaultPinPath
	seen := make(map[string]bool)
	for _, iface := range ifaces {
		path := filepath.Join(pinBase, fmt.Sprintf("neigh-%s", iface))
		if seen[path] {
			t.Errorf("duplicate pin path: %s", path)
		}
		seen[path] = true
	}
}

func TestDefaultPinPath(t *testing.T) {
	if DefaultPinPath != "/sys/fs/bpf/l2radar" {
		t.Errorf("unexpected default pin path: %s", DefaultPinPath)
	}
}

func TestMapPinPermissions(t *testing.T) {
	if MapPinPermissions != os.FileMode(0444) {
		t.Errorf("unexpected map permissions: %o", MapPinPermissions)
	}
}

func TestAttachInvalidInterface(t *testing.T) {
	_, err := Attach("nonexistent_iface_12345", DefaultPinPath, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent interface")
	}
}

func TestAttachAndClose(t *testing.T) {
	// This test requires root/CAP_BPF + a real interface.
	// It validates the full attach/close lifecycle.
	iface := os.Getenv("L2RADAR_TEST_IFACE")
	if iface == "" {
		t.Skip("set L2RADAR_TEST_IFACE to run this test")
	}

	tmpDir := t.TempDir()
	pinBase := filepath.Join(tmpDir, "l2radar")

	probe, err := Attach(iface, pinBase, nil)
	if err != nil {
		t.Skipf("cannot attach (likely missing privileges): %v", err)
	}
	defer probe.Close()

	if probe.Interface() != iface {
		t.Errorf("expected interface %s, got %s", iface, probe.Interface())
	}

	expectedPin := filepath.Join(pinBase, fmt.Sprintf("neigh-%s", iface))
	if probe.MapPinPath() != expectedPin {
		t.Errorf("expected pin path %s, got %s", expectedPin, probe.MapPinPath())
	}

	// Verify pin file exists and has correct permissions
	info, err := os.Stat(probe.MapPinPath())
	if err != nil {
		t.Fatalf("pin file not found: %v", err)
	}
	if info.Mode().Perm() != MapPinPermissions {
		t.Errorf("expected permissions %o, got %o", MapPinPermissions, info.Mode().Perm())
	}

	// Close should clean up
	if err := probe.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	// Pin should be removed
	if _, err := os.Stat(expectedPin); !os.IsNotExist(err) {
		t.Error("pin file should be removed after close")
	}
}
