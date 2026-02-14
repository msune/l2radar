package oui

import (
	"net"
	"testing"
)

func TestLookupKnownVendor(t *testing.T) {
	// 28:6F:B9 is Nokia Shanghai Bell Co., Ltd. (first entry in oui.txt)
	mac, _ := net.ParseMAC("28:6F:B9:00:00:00")
	vendor := Lookup(mac)
	if vendor == "" {
		t.Fatal("expected non-empty vendor for 28:6F:B9")
	}
	if vendor != "Nokia Shanghai Bell Co., Ltd." {
		t.Errorf("expected 'Nokia Shanghai Bell Co., Ltd.', got %q", vendor)
	}
}

func TestLookupAnotherVendor(t *testing.T) {
	// 08:EA:44 is Extreme Networks Headquarters
	mac, _ := net.ParseMAC("08:EA:44:11:22:33")
	vendor := Lookup(mac)
	if vendor != "Extreme Networks Headquarters" {
		t.Errorf("expected 'Extreme Networks Headquarters', got %q", vendor)
	}
}

func TestLookupUnknownVendor(t *testing.T) {
	// FF:FF:FF is broadcast, unlikely to have an OUI entry
	mac, _ := net.ParseMAC("FF:FF:FF:00:00:00")
	vendor := Lookup(mac)
	if vendor != "" {
		t.Errorf("expected empty string for unknown OUI, got %q", vendor)
	}
}

func TestLookupNilMAC(t *testing.T) {
	vendor := Lookup(nil)
	if vendor != "" {
		t.Errorf("expected empty string for nil MAC, got %q", vendor)
	}
}

func TestLookupShortMAC(t *testing.T) {
	vendor := Lookup(net.HardwareAddr{0x28, 0x6F})
	if vendor != "" {
		t.Errorf("expected empty string for short MAC, got %q", vendor)
	}
}

func TestDatabaseLoaded(t *testing.T) {
	// The embedded database should have a substantial number of entries
	count := EntryCount()
	if count < 1000 {
		t.Errorf("expected at least 1000 OUI entries, got %d", count)
	}
}
