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
			// 28:6F:B9 = Nokia Shanghai Bell Co., Ltd.
			MAC:       net.HardwareAddr{0x28, 0x6f, 0xb9, 0x11, 0x00, 0x01},
			IPv4:      []net.IP{net.ParseIP("192.168.1.1").To4()},
			IPv6:      []net.IP{net.ParseIP("fe80::1")},
			FirstSeen: earlier,
			LastSeen:  now,
		},
		{
			// Locally administered MAC — no OUI entry
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
	if !strings.Contains(output, "28:6f:b9:11:00:01") {
		t.Error("table should contain first MAC")
	}
	if !strings.Contains(output, "02:42:ac:11:00:02") {
		t.Error("table should contain second MAC")
	}

	// Should contain vendor name for known OUI
	if !strings.Contains(output, "(Nokia Shanghai Bell") {
		t.Errorf("table should contain vendor name for Nokia OUI, got:\n%s", output)
	}

	// Unknown OUI should not have parentheses
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "02:42:ac:11:00:02") {
			if strings.Contains(line, "(") {
				t.Error("unknown OUI should not have vendor in parentheses")
			}
			break
		}
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

func TestKtimeToTimeBasic(t *testing.T) {
	// Simulate: wall clock is 14:30:00, system has been up for 1 hour.
	// An event at ktime=30min should map to 13:30 + 30min = 14:00.
	wallTime := time.Date(2026, 2, 14, 14, 30, 0, 0, time.UTC)
	uptimeNs := int64(1 * 60 * 60 * 1e9) // 1 hour

	origTimeNow := timeNow
	origMonoNow := monoNow
	defer func() { timeNow = origTimeNow; monoNow = origMonoNow }()

	timeNow = func() time.Time { return wallTime }
	monoNow = func() int64 { return uptimeNs }

	thirtyMinNs := uint64(30 * 60 * 1e9)
	got := ktimeToTime(thirtyMinNs)
	expected := time.Date(2026, 2, 14, 14, 0, 0, 0, time.UTC)

	if got.Sub(expected).Abs() > time.Second {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestKtimeToTimeZero(t *testing.T) {
	got := ktimeToTime(0)
	if !got.IsZero() {
		t.Errorf("expected zero time for ktime=0, got %v", got)
	}
}

func TestKtimeToTimeStableAcrossRuns(t *testing.T) {
	// Simulate two dump invocations 10 minutes apart.
	// The same ktime value must produce the same wall-clock time.
	origTimeNow := timeNow
	origMonoNow := monoNow
	defer func() { timeNow = origTimeNow; monoNow = origMonoNow }()

	firstSeenKtime := uint64(1800 * 1e9) // 30 min after boot

	// First dump: system up 1 hour, wall clock 14:00
	timeNow = func() time.Time { return time.Date(2026, 2, 14, 14, 0, 0, 0, time.UTC) }
	monoNow = func() int64 { return int64(3600 * 1e9) }
	first := ktimeToTime(firstSeenKtime)

	// Second dump: system up 1h10m, wall clock 14:10
	timeNow = func() time.Time { return time.Date(2026, 2, 14, 14, 10, 0, 0, time.UTC) }
	monoNow = func() int64 { return int64(4200 * 1e9) }
	second := ktimeToTime(firstSeenKtime)

	if first.Sub(second).Abs() > time.Second {
		t.Errorf("first_seen should be stable across runs: first=%v, second=%v", first, second)
	}
}

func TestKtimeToTimeFirstAndLastSeenDiffer(t *testing.T) {
	// first_seen and last_seen with different ktime values must produce
	// different wall-clock times.
	origTimeNow := timeNow
	origMonoNow := monoNow
	defer func() { timeNow = origTimeNow; monoNow = origMonoNow }()

	timeNow = func() time.Time { return time.Date(2026, 2, 14, 14, 30, 0, 0, time.UTC) }
	monoNow = func() int64 { return int64(3600 * 1e9) }

	firstSeenKtime := uint64(1800 * 1e9)  // 30 min after boot
	lastSeenKtime := uint64(3500 * 1e9)   // ~58 min after boot

	firstSeen := ktimeToTime(firstSeenKtime)
	lastSeen := ktimeToTime(lastSeenKtime)

	diff := lastSeen.Sub(firstSeen)
	expectedDiff := time.Duration(lastSeenKtime-firstSeenKtime) * time.Nanosecond

	if (diff - expectedDiff).Abs() > time.Second {
		t.Errorf("difference between first/last seen should be %v, got %v", expectedDiff, diff)
	}
}
