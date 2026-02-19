package cli

import (
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/marc/l2radar/probe/pkg/dump"
	"github.com/marc/l2radar/probe/pkg/export"
)

func TestMarshalDumpJSONIncludesInterfaceInfoAndStats(t *testing.T) {
	ts := time.Date(2026, 2, 14, 14, 30, 0, 0, time.UTC)
	iface := "lo"
	expectMAC := false
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, candidate := range ifaces {
			if len(candidate.HardwareAddr) == 0 {
				continue
			}
			iface = candidate.Name
			expectMAC = true
			break
		}
	}

	neighbours := []dump.Neighbour{
		{
			MAC:       net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			IPv4:      []net.IP{net.ParseIP("127.0.0.2").To4()},
			IPv6:      []net.IP{net.ParseIP("::1")},
			FirstSeen: ts.Add(-time.Minute),
			LastSeen:  ts,
		},
	}

	b, err := marshalDumpJSON(iface, ts, neighbours)
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("skipping due to restricted netlink access: %v", err)
		}
		t.Fatalf("marshalDumpJSON failed: %v", err)
	}

	var parsed export.InterfaceData
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.Interface != iface {
		t.Fatalf("expected interface %s, got %q", iface, parsed.Interface)
	}
	if parsed.ExportInterval != "0s" {
		t.Fatalf("expected export_interval 0s, got %q", parsed.ExportInterval)
	}
	if expectMAC && parsed.MAC == "" {
		t.Fatal("expected interface MAC to be populated")
	}
	if parsed.Stats == nil {
		t.Fatal("expected stats to be populated")
	}
	if len(parsed.Neighbours) != 1 {
		t.Fatalf("expected 1 neighbour, got %d", len(parsed.Neighbours))
	}
}

func TestMarshalDumpJSONLookupError(t *testing.T) {
	_, err := marshalDumpJSON("definitely-not-an-interface", time.Now(), nil)
	if err == nil {
		t.Fatal("expected error for unknown interface")
	}
}
