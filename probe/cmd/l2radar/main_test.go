package main

import (
	"net"
	"testing"
)

func TestStringSliceSet(t *testing.T) {
	var s stringSlice
	if err := s.Set("eth0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Set("wlan0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(s))
	}
	if s[0] != "eth0" || s[1] != "wlan0" {
		t.Errorf("unexpected values: %v", s)
	}
}

func TestStringSliceString(t *testing.T) {
	s := stringSlice{"eth0", "wlan0"}
	str := s.String()
	if str != "eth0, wlan0" {
		t.Errorf("expected 'eth0, wlan0', got '%s'", str)
	}
}

func TestStringSliceEmpty(t *testing.T) {
	var s stringSlice
	if s.String() != "" {
		t.Errorf("expected empty string, got '%s'", s.String())
	}
}

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

func TestResolveInterfaces_AnyCaseInsensitive(t *testing.T) {
	original := listInterfaces
	defer func() { listInterfaces = original }()

	listInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{
			{Name: "eth0", Flags: net.FlagUp},
		}, nil
	}

	for _, name := range []string{"any", "ANY", "Any"} {
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
