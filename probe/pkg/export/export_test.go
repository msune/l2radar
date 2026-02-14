package export

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marc/l2radar/probe/pkg/dump"
)

func TestInterfaceDataJSON(t *testing.T) {
	now := time.Date(2026, 2, 14, 14, 30, 0, 0, time.UTC)
	neighbours := []dump.Neighbour{
		{
			MAC:       net.HardwareAddr{0xdc, 0x4b, 0xa1, 0x69, 0x38, 0x16},
			IPv4:      []net.IP{net.ParseIP("192.168.1.33").To4(), net.ParseIP("10.0.0.5").To4()},
			IPv6:      []net.IP{net.ParseIP("fe80::c09e:74a6:4353:cd6a"), net.ParseIP("2001:db8::1")},
			FirstSeen: time.Date(2026, 2, 14, 14, 19, 44, 0, time.UTC),
			LastSeen:  time.Date(2026, 2, 14, 14, 32, 18, 0, time.UTC),
		},
	}

	data := NewInterfaceData("eth0", now, neighbours)
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Verify it round-trips through JSON correctly
	var parsed InterfaceData
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.Interface != "eth0" {
		t.Errorf("expected interface eth0, got %s", parsed.Interface)
	}
	if parsed.Timestamp != "2026-02-14T14:30:00Z" {
		t.Errorf("expected timestamp 2026-02-14T14:30:00Z, got %s", parsed.Timestamp)
	}
	if len(parsed.Neighbours) != 1 {
		t.Fatalf("expected 1 neighbour, got %d", len(parsed.Neighbours))
	}
	n := parsed.Neighbours[0]
	if n.MAC != "dc:4b:a1:69:38:16" {
		t.Errorf("expected MAC dc:4b:a1:69:38:16, got %s", n.MAC)
	}
	if len(n.IPv4) != 2 || n.IPv4[0] != "192.168.1.33" || n.IPv4[1] != "10.0.0.5" {
		t.Errorf("unexpected IPv4: %v", n.IPv4)
	}
	if len(n.IPv6) != 2 || n.IPv6[0] != "fe80::c09e:74a6:4353:cd6a" || n.IPv6[1] != "2001:db8::1" {
		t.Errorf("unexpected IPv6: %v", n.IPv6)
	}
	if n.FirstSeen != "2026-02-14T14:19:44Z" {
		t.Errorf("expected first_seen 2026-02-14T14:19:44Z, got %s", n.FirstSeen)
	}
	if n.LastSeen != "2026-02-14T14:32:18Z" {
		t.Errorf("expected last_seen 2026-02-14T14:32:18Z, got %s", n.LastSeen)
	}
}

func TestEmptyNeighbours(t *testing.T) {
	now := time.Date(2026, 2, 14, 14, 0, 0, 0, time.UTC)
	data := NewInterfaceData("eth0", now, nil)

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed InterfaceData
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.Neighbours == nil {
		t.Error("neighbours should be empty array, not null")
	}
	if len(parsed.Neighbours) != 0 {
		t.Errorf("expected 0 neighbours, got %d", len(parsed.Neighbours))
	}
}

func TestNeighbourNoIPs(t *testing.T) {
	now := time.Now()
	neighbours := []dump.Neighbour{
		{
			MAC:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0x02},
			IPv4:      nil,
			IPv6:      nil,
			FirstSeen: now,
			LastSeen:  now,
		},
	}

	data := NewInterfaceData("eth0", now, neighbours)
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed InterfaceData
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	n := parsed.Neighbours[0]
	if n.IPv4 == nil {
		t.Error("ipv4 should be empty array, not null")
	}
	if n.IPv6 == nil {
		t.Error("ipv6 should be empty array, not null")
	}
}

func TestTimestampFormatRFC3339(t *testing.T) {
	ts := time.Date(2026, 2, 14, 14, 30, 0, 0, time.UTC)
	neighbours := []dump.Neighbour{
		{
			MAC:       net.HardwareAddr{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			FirstSeen: ts,
			LastSeen:  ts,
		},
	}

	data := NewInterfaceData("eth0", ts, neighbours)
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed InterfaceData
	json.Unmarshal(b, &parsed)

	// All timestamps must parse as valid RFC3339
	if _, err := time.Parse(time.RFC3339, parsed.Timestamp); err != nil {
		t.Errorf("timestamp not RFC3339: %s", parsed.Timestamp)
	}
	if _, err := time.Parse(time.RFC3339, parsed.Neighbours[0].FirstSeen); err != nil {
		t.Errorf("first_seen not RFC3339: %s", parsed.Neighbours[0].FirstSeen)
	}
	if _, err := time.Parse(time.RFC3339, parsed.Neighbours[0].LastSeen); err != nil {
		t.Errorf("last_seen not RFC3339: %s", parsed.Neighbours[0].LastSeen)
	}
}

func TestWriteJSONAtomicCreatesFile(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 2, 14, 14, 30, 0, 0, time.UTC)
	neighbours := []dump.Neighbour{
		{
			MAC:       net.HardwareAddr{0xdc, 0x4b, 0xa1, 0x69, 0x38, 0x16},
			IPv4:      []net.IP{net.ParseIP("192.168.1.33").To4()},
			IPv6:      []net.IP{net.ParseIP("fe80::1")},
			FirstSeen: now.Add(-10 * time.Minute),
			LastSeen:  now,
		},
	}

	err := WriteJSON("eth0", neighbours, dir, now)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	outPath := filepath.Join(dir, "neigh-eth0.json")
	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}

	var parsed InterfaceData
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.Interface != "eth0" {
		t.Errorf("expected interface eth0, got %s", parsed.Interface)
	}
	if len(parsed.Neighbours) != 1 {
		t.Errorf("expected 1 neighbour, got %d", len(parsed.Neighbours))
	}
}

func TestWriteJSONOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	// Write first version
	err := WriteJSON("eth0", nil, dir, now)
	if err != nil {
		t.Fatalf("first WriteJSON failed: %v", err)
	}

	// Write second version with a neighbour
	neighbours := []dump.Neighbour{
		{
			MAC:       net.HardwareAddr{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			FirstSeen: now,
			LastSeen:  now,
		},
	}
	err = WriteJSON("eth0", neighbours, dir, now)
	if err != nil {
		t.Fatalf("second WriteJSON failed: %v", err)
	}

	outPath := filepath.Join(dir, "neigh-eth0.json")
	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}

	var parsed InterfaceData
	json.Unmarshal(b, &parsed)
	if len(parsed.Neighbours) != 1 {
		t.Errorf("expected 1 neighbour after overwrite, got %d", len(parsed.Neighbours))
	}
}

func TestOutputFileName(t *testing.T) {
	if got := OutputFileName("eth0"); got != "neigh-eth0.json" {
		t.Errorf("expected neigh-eth0.json, got %s", got)
	}
	if got := OutputFileName("wlan0"); got != "neigh-wlan0.json" {
		t.Errorf("expected neigh-wlan0.json, got %s", got)
	}
}

// TestGoldenFileSchema validates that our JSON output matches the golden
// file schema used as a contract between probe and UI.
func TestGoldenFileSchema(t *testing.T) {
	goldenDir := filepath.Join("..", "..", "..", "testdata")

	for _, tc := range []struct {
		file  string
		iface string
	}{
		{"neigh-eth0.json", "eth0"},
		{"neigh-wlan0.json", "wlan0"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join(goldenDir, tc.file))
			if err != nil {
				t.Fatalf("reading golden file: %v", err)
			}

			var data InterfaceData
			if err := json.Unmarshal(b, &data); err != nil {
				t.Fatalf("golden file does not parse as InterfaceData: %v", err)
			}

			// Validate top-level fields
			if data.Interface != tc.iface {
				t.Errorf("expected interface %s, got %s", tc.iface, data.Interface)
			}
			if _, err := time.Parse(time.RFC3339, data.Timestamp); err != nil {
				t.Errorf("timestamp not RFC3339: %s", data.Timestamp)
			}

			// Validate each neighbour
			for i, n := range data.Neighbours {
				// MAC format: xx:xx:xx:xx:xx:xx
				if _, err := net.ParseMAC(n.MAC); err != nil {
					t.Errorf("neighbour[%d]: invalid MAC %q: %v", i, n.MAC, err)
				}

				// IPv4 must be valid
				for _, ip := range n.IPv4 {
					if net.ParseIP(ip) == nil {
						t.Errorf("neighbour[%d]: invalid IPv4 %q", i, ip)
					}
				}

				// IPv6 must be valid
				for _, ip := range n.IPv6 {
					if net.ParseIP(ip) == nil {
						t.Errorf("neighbour[%d]: invalid IPv6 %q", i, ip)
					}
				}

				// Timestamps must be RFC3339
				if _, err := time.Parse(time.RFC3339, n.FirstSeen); err != nil {
					t.Errorf("neighbour[%d]: first_seen not RFC3339: %s", i, n.FirstSeen)
				}
				if _, err := time.Parse(time.RFC3339, n.LastSeen); err != nil {
					t.Errorf("neighbour[%d]: last_seen not RFC3339: %s", i, n.LastSeen)
				}

				// ipv4 and ipv6 must never be null
				if n.IPv4 == nil {
					t.Errorf("neighbour[%d]: ipv4 is null", i)
				}
				if n.IPv6 == nil {
					t.Errorf("neighbour[%d]: ipv6 is null", i)
				}
			}
		})
	}
}
