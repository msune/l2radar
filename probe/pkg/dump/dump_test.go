package dump

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestFormatMAC(t *testing.T) {
	mac := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x02}
	got := mac.String()
	if got != "02:42:ac:11:00:02" {
		t.Errorf("unexpected MAC format: %s", got)
	}
}

func TestNeighbourIPv4Strings(t *testing.T) {
	n := Neighbour{
		MAC:  net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x02},
		IPv4: []net.IP{net.ParseIP("192.168.1.1").To4(), net.ParseIP("10.0.0.1").To4()},
	}
	got := n.IPv4String()
	if got != "192.168.1.1, 10.0.0.1" {
		t.Errorf("unexpected IPv4 string: %s", got)
	}
}

func TestNeighbourIPv4StringEmpty(t *testing.T) {
	n := Neighbour{}
	if n.IPv4String() != "" {
		t.Errorf("expected empty string, got %s", n.IPv4String())
	}
}

func TestNeighbourIPv6Strings(t *testing.T) {
	n := Neighbour{
		IPv6: []net.IP{net.ParseIP("fe80::1"), net.ParseIP("2001:db8::1")},
	}
	got := n.IPv6String()
	if got != "fe80::1, 2001:db8::1" {
		t.Errorf("unexpected IPv6 string: %s", got)
	}
}

func TestFormatTable(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-5 * time.Minute)

	neighbours := []Neighbour{
		{
			MAC:       net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x01},
			IPv4:      []net.IP{net.ParseIP("192.168.1.1").To4()},
			IPv6:      []net.IP{net.ParseIP("fe80::1")},
			FirstSeen: earlier,
			LastSeen:  now,
		},
		{
			MAC:       net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x02},
			IPv4:      nil,
			IPv6:      nil,
			FirstSeen: earlier,
			LastSeen:  earlier,
		},
	}

	var buf strings.Builder
	FormatTable(&buf, neighbours)
	output := buf.String()

	// Should contain header
	if !strings.Contains(output, "MAC") {
		t.Error("table should contain MAC header")
	}
	if !strings.Contains(output, "IPv4") {
		t.Error("table should contain IPv4 header")
	}
	if !strings.Contains(output, "IPv6") {
		t.Error("table should contain IPv6 header")
	}
	if !strings.Contains(output, "FIRST SEEN") {
		t.Error("table should contain FIRST SEEN header")
	}
	if !strings.Contains(output, "LAST SEEN") {
		t.Error("table should contain LAST SEEN header")
	}

	// Should contain our MACs
	if !strings.Contains(output, "02:42:ac:11:00:01") {
		t.Error("table should contain first MAC")
	}
	if !strings.Contains(output, "02:42:ac:11:00:02") {
		t.Error("table should contain second MAC")
	}

	// Should contain IP
	if !strings.Contains(output, "192.168.1.1") {
		t.Error("table should contain IPv4 address")
	}
	if !strings.Contains(output, "fe80::1") {
		t.Error("table should contain IPv6 address")
	}
}

func TestSortByLastSeen(t *testing.T) {
	now := time.Now()
	old := now.Add(-10 * time.Minute)

	neighbours := []Neighbour{
		{MAC: net.HardwareAddr{0x01}, LastSeen: old},
		{MAC: net.HardwareAddr{0x02}, LastSeen: now},
		{MAC: net.HardwareAddr{0x03}, LastSeen: now.Add(-5 * time.Minute)},
	}

	SortByLastSeen(neighbours)

	if neighbours[0].MAC[0] != 0x02 {
		t.Errorf("expected most recent first, got MAC %x", neighbours[0].MAC[0])
	}
	if neighbours[1].MAC[0] != 0x03 {
		t.Errorf("expected second most recent second, got MAC %x", neighbours[1].MAC[0])
	}
	if neighbours[2].MAC[0] != 0x01 {
		t.Errorf("expected oldest last, got MAC %x", neighbours[2].MAC[0])
	}
}

func TestFormatTableEmpty(t *testing.T) {
	var buf strings.Builder
	FormatTable(&buf, nil)
	output := buf.String()

	// Should still have a header
	if !strings.Contains(output, "MAC") {
		t.Error("empty table should still have header")
	}
}

func TestPinPath(t *testing.T) {
	path := PinPath("/sys/fs/bpf/l2radar", "eth0")
	if path != "/sys/fs/bpf/l2radar/neigh-eth0" {
		t.Errorf("unexpected pin path: %s", path)
	}
}
