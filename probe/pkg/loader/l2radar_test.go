package loader

import (
	"encoding/binary"
	"errors"
	"net"
	"os"
	"testing"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/rlimit"
)

func TestMain(m *testing.M) {
	// Remove memlock rlimit for eBPF operations.
	// Ignore errors — may lack permissions.
	_ = rlimit.RemoveMemlock()
	os.Exit(m.Run())
}

// buildEthernetFrame constructs a minimal Ethernet frame.
func buildEthernetFrame(dst, src net.HardwareAddr, etherType uint16, payload []byte) []byte {
	frame := make([]byte, 14+len(payload))
	copy(frame[0:6], dst)
	copy(frame[6:12], src)
	binary.BigEndian.PutUint16(frame[12:14], etherType)
	copy(frame[14:], payload)
	return frame
}

// macKey constructs a mac_key struct matching the eBPF map key layout.
func macKey(mac net.HardwareAddr) l2radarMacKey {
	var key l2radarMacKey
	copy(key.Addr[:], mac)
	return key
}

// loadTestObjects loads the eBPF program and maps for testing.
// Skips the test if running without sufficient privileges.
func loadTestObjects(t *testing.T) (*l2radarObjects, func()) {
	t.Helper()
	var objs l2radarObjects
	err := loadL2radarObjects(&objs, &ebpf.CollectionOptions{})
	if err != nil {
		if errors.Is(err, os.ErrPermission) || errors.Is(err, ebpf.ErrNotSupported) {
			t.Skip("skipping: insufficient privileges to load eBPF programs")
		}
		t.Fatalf("loading eBPF objects: %v", err)
	}
	return &objs, func() { objs.Close() }
}

// runProgram runs the eBPF program with the given packet and returns the
// return code.
func runProgram(t *testing.T, prog *ebpf.Program, pkt []byte) uint32 {
	t.Helper()
	ret, _, err := prog.Test(pkt)
	if err != nil {
		t.Fatalf("running eBPF program: %v", err)
	}
	return ret
}

// lookupNeighbour looks up a MAC in the neighbours map.
func lookupNeighbour(t *testing.T, m *ebpf.Map, mac net.HardwareAddr) (*l2radarNeighbourEntry, bool) {
	t.Helper()
	key := macKey(mac)
	var val l2radarNeighbourEntry
	err := m.Lookup(&key, &val)
	if err != nil {
		return nil, false
	}
	return &val, true
}

// --- Tests ---

func TestMulticastMACNotTracked(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	// Multicast MAC: bit 0 of first byte is set
	multicastMAC := net.HardwareAddr{0x01, 0x00, 0x5e, 0x00, 0x00, 0x01}
	dstMAC := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}

	pkt := buildEthernetFrame(dstMAC, multicastMAC, 0x0800, make([]byte, 46))
	ret := runProgram(t, objs.L2radar, pkt)

	if ret != 0xffffffff { // TC_ACT_UNSPEC = -1, as uint32 = 0xffffffff
		t.Errorf("expected TC_ACT_UNSPEC (-1), got %d", ret)
	}

	if _, found := lookupNeighbour(t, objs.Neighbours, multicastMAC); found {
		t.Error("multicast MAC should not be tracked")
	}
}

func TestBroadcastMACNotTracked(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	broadcastMAC := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	dstMAC := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}

	pkt := buildEthernetFrame(dstMAC, broadcastMAC, 0x0800, make([]byte, 46))
	runProgram(t, objs.L2radar, pkt)

	if _, found := lookupNeighbour(t, objs.Neighbours, broadcastMAC); found {
		t.Error("broadcast MAC should not be tracked")
	}
}

func TestUnicastMACTracked(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	srcMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x02}
	dstMAC := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}

	pkt := buildEthernetFrame(dstMAC, srcMAC, 0x0800, make([]byte, 46))
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, srcMAC)
	if !found {
		t.Fatal("unicast MAC should be tracked")
	}

	if entry.Ipv4Count != 0 {
		t.Errorf("expected ipv4_count=0, got %d", entry.Ipv4Count)
	}
	if entry.Ipv6Count != 0 {
		t.Errorf("expected ipv6_count=0, got %d", entry.Ipv6Count)
	}
	if entry.FirstSeen == 0 {
		t.Error("first_seen should be set")
	}
	if entry.LastSeen == 0 {
		t.Error("last_seen should be set")
	}
}

func TestTimestampsFirstSeenStable(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	srcMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x03}
	dstMAC := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}

	pkt := buildEthernetFrame(dstMAC, srcMAC, 0x0800, make([]byte, 46))

	// First packet
	runProgram(t, objs.L2radar, pkt)
	entry1, found := lookupNeighbour(t, objs.Neighbours, srcMAC)
	if !found {
		t.Fatal("MAC not found after first packet")
	}
	firstSeen1 := entry1.FirstSeen
	lastSeen1 := entry1.LastSeen

	// Second packet
	runProgram(t, objs.L2radar, pkt)
	entry2, found := lookupNeighbour(t, objs.Neighbours, srcMAC)
	if !found {
		t.Fatal("MAC not found after second packet")
	}

	if entry2.FirstSeen != firstSeen1 {
		t.Errorf("first_seen changed: %d -> %d", firstSeen1, entry2.FirstSeen)
	}
	if entry2.LastSeen < lastSeen1 {
		t.Errorf("last_seen should not decrease: %d -> %d", lastSeen1, entry2.LastSeen)
	}
}

func TestReturnValueAlwaysUnspec(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	cases := []struct {
		name string
		src  net.HardwareAddr
	}{
		{"unicast", net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x04}},
		{"multicast", net.HardwareAddr{0x01, 0x00, 0x5e, 0x00, 0x00, 0x01}},
		{"broadcast", net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}
	dstMAC := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pkt := buildEthernetFrame(dstMAC, tc.src, 0x0800, make([]byte, 46))
			ret := runProgram(t, objs.L2radar, pkt)
			// TC_ACT_UNSPEC is -1, represented as 0xffffffff in uint32
			if ret != 0xffffffff {
				t.Errorf("expected TC_ACT_UNSPEC (0xffffffff), got 0x%x", ret)
			}
		})
	}
}
