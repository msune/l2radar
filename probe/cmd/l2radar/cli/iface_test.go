package cli

import (
	"net"
	"testing"
)

func TestResolveInterfaces_ExplicitNames(t *testing.T) {
	result, err := resolveInterfaces([]string{"eth0", "wlan0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0] != "eth0" || result[1] != "wlan0" {
		t.Fatalf("expected [eth0 wlan0], got %v", result)
	}
}

func TestResolveInterfaces_Any(t *testing.T) {
	original := listInterfaces
	defer func() { listInterfaces = original }()

	listInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{Name: "lo", Flags: net.FlagLoopback | net.FlagUp},
			{Name: "eth0", Flags: net.FlagUp},
			{Name: "wlan0", Flags: net.FlagUp},
			{Name: "docker0", Flags: net.FlagUp},
		}, nil
	}

	result, err := resolveInterfaces([]string{"any"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "any" includes everything except loopback (including docker0)
	expected := []string{"eth0", "wlan0", "docker0"}
	if len(result) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
	for i, name := range expected {
		if result[i] != name {
			t.Fatalf("expected %v at index %d, got %v", name, i, result[i])
		}
	}
}

func TestResolveInterfaces_External(t *testing.T) {
	original := listInterfaces
	defer func() { listInterfaces = original }()

	listInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{Name: "lo", Flags: net.FlagLoopback | net.FlagUp},
			{Name: "eth0", Flags: net.FlagUp},
			{Name: "wlan0", Flags: net.FlagUp},
			{Name: "docker0", Flags: net.FlagUp},
			{Name: "veth1234", Flags: net.FlagUp},
			{Name: "br-abcdef", Flags: net.FlagUp},
			{Name: "virbr0", Flags: net.FlagUp},
		}, nil
	}

	result, err := resolveInterfaces([]string{"external"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "external" filters out virtual interfaces (docker*, veth*, br-*, virbr*)
	expected := []string{"eth0", "wlan0"}
	if len(result) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
	for i, name := range expected {
		if result[i] != name {
			t.Fatalf("expected %v at index %d, got %v", name, i, result[i])
		}
	}
}

func TestResolveInterfaces_AnySkipsLoopback(t *testing.T) {
	original := listInterfaces
	defer func() { listInterfaces = original }()

	listInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{Name: "lo", Flags: net.FlagLoopback | net.FlagUp},
		}, nil
	}

	result, err := resolveInterfaces([]string{"any"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}

func TestResolveInterfaces_ExternalSkipsVirtual(t *testing.T) {
	original := listInterfaces
	defer func() { listInterfaces = original }()

	listInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{Name: "lo", Flags: net.FlagLoopback | net.FlagUp},
			{Name: "docker0", Flags: net.FlagUp},
			{Name: "vethabcdef", Flags: net.FlagUp},
			{Name: "br-12345", Flags: net.FlagUp},
			{Name: "virbr0", Flags: net.FlagUp},
		}, nil
	}

	result, err := resolveInterfaces([]string{"external"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty (all virtual), got %v", result)
	}
}

func TestResolveInterfaces_CaseInsensitive(t *testing.T) {
	original := listInterfaces
	defer func() { listInterfaces = original }()

	listInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{Name: "eth0", Flags: net.FlagUp},
		}, nil
	}

	for _, name := range []string{"any", "ANY", "Any", "external", "EXTERNAL", "External"} {
		result, err := resolveInterfaces([]string{name})
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", name, err)
		}
		if len(result) != 1 || result[0] != "eth0" {
			t.Fatalf("expected [eth0] for %q, got %v", name, result)
		}
	}
}

func TestResolveInterfaces_Dedup(t *testing.T) {
	original := listInterfaces
	defer func() { listInterfaces = original }()

	listInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{Name: "lo", Flags: net.FlagLoopback | net.FlagUp},
			{Name: "eth0", Flags: net.FlagUp},
			{Name: "wlan0", Flags: net.FlagUp},
		}, nil
	}

	result, err := resolveInterfaces([]string{"eth0", "any"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"eth0", "wlan0"}
	if len(result) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
	for i, name := range expected {
		if result[i] != name {
			t.Fatalf("expected %v at index %d, got %v", name, i, result[i])
		}
	}
}

func TestResolveInterfaces_MixedAnyAndExplicit(t *testing.T) {
	original := listInterfaces
	defer func() { listInterfaces = original }()

	listInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{Name: "lo", Flags: net.FlagLoopback | net.FlagUp},
			{Name: "eth0", Flags: net.FlagUp},
		}, nil
	}

	result, err := resolveInterfaces([]string{"br0", "any"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"br0", "eth0"}
	if len(result) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
	for i, name := range expected {
		if result[i] != name {
			t.Fatalf("expected %v at index %d, got %v", name, i, result[i])
		}
	}
}

func TestResolveInterfaces_MixedExternalAndExplicit(t *testing.T) {
	original := listInterfaces
	defer func() { listInterfaces = original }()

	listInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{Name: "lo", Flags: net.FlagLoopback | net.FlagUp},
			{Name: "eth0", Flags: net.FlagUp},
			{Name: "docker0", Flags: net.FlagUp},
		}, nil
	}

	// "external" excludes docker0; explicit "docker0" can still be added separately
	result, err := resolveInterfaces([]string{"docker0", "external"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"docker0", "eth0"}
	if len(result) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
	for i, name := range expected {
		if result[i] != name {
			t.Fatalf("expected %v at index %d, got %v", name, i, result[i])
		}
	}
}
