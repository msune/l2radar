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

// buildVLANEthernetFrame constructs an 802.1Q-tagged Ethernet frame.
func buildVLANEthernetFrame(dst, src net.HardwareAddr, vlanID uint16, etherType uint16, payload []byte) []byte {
	frame := make([]byte, 18+len(payload))
	copy(frame[0:6], dst)
	copy(frame[6:12], src)
	binary.BigEndian.PutUint16(frame[12:14], 0x8100) // TPID
	binary.BigEndian.PutUint16(frame[14:16], vlanID)  // TCI (PCP=0, DEI=0, VID)
	binary.BigEndian.PutUint16(frame[16:18], etherType)
	copy(frame[18:], payload)
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

// buildARPPacket constructs an ARP packet with Ethernet header.
// opcode: 1=request, 2=reply
func buildARPPacket(ethDst, ethSrc net.HardwareAddr, opcode uint16,
	senderMAC net.HardwareAddr, senderIP net.IP,
	targetMAC net.HardwareAddr, targetIP net.IP) []byte {

	// ARP payload: 28 bytes for IPv4-over-Ethernet
	arp := make([]byte, 28)
	binary.BigEndian.PutUint16(arp[0:2], 1)      // htype = Ethernet
	binary.BigEndian.PutUint16(arp[2:4], 0x0800)  // ptype = IPv4
	arp[4] = 6                                     // hlen
	arp[5] = 4                                     // plen
	binary.BigEndian.PutUint16(arp[6:8], opcode)
	copy(arp[8:14], senderMAC)
	copy(arp[14:18], senderIP.To4())
	copy(arp[18:24], targetMAC)
	copy(arp[24:28], targetIP.To4())

	return buildEthernetFrame(ethDst, ethSrc, 0x0806, arp)
}

// ipv4FromEntry extracts the active IPv4 addresses from a neighbour entry.
func ipv4FromEntry(entry *l2radarNeighbourEntry) []net.IP {
	var ips []net.IP
	for i := 0; i < int(entry.Ipv4Count); i++ {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, entry.Ipv4[i])
		ips = append(ips, ip)
	}
	return ips
}

// containsIPv4 checks if an IP is in the list.
func containsIPv4(ips []net.IP, target net.IP) bool {
	for _, ip := range ips {
		if ip.Equal(target) {
			return true
		}
	}
	return false
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

// --- ARP Tests ---

func TestARPRequestSenderTracked(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	senderMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x10}
	senderIP := net.ParseIP("192.168.1.10").To4()
	targetMAC := net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	targetIP := net.ParseIP("192.168.1.1").To4()
	broadcast := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	pkt := buildARPPacket(broadcast, senderMAC, 1, senderMAC, senderIP, targetMAC, targetIP)
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, senderMAC)
	if !found {
		t.Fatal("sender MAC should be tracked from ARP request")
	}
	if entry.Ipv4Count != 1 {
		t.Fatalf("expected ipv4_count=1, got %d", entry.Ipv4Count)
	}
	ips := ipv4FromEntry(entry)
	if !containsIPv4(ips, senderIP) {
		t.Errorf("sender IP %s not found in entry", senderIP)
	}
}

func TestARPReplySenderAndTargetTracked(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	senderMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x20}
	senderIP := net.ParseIP("192.168.1.1").To4()
	targetMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x21}
	targetIP := net.ParseIP("192.168.1.10").To4()

	pkt := buildARPPacket(targetMAC, senderMAC, 2, senderMAC, senderIP, targetMAC, targetIP)
	runProgram(t, objs.L2radar, pkt)

	// Sender should have its IP tracked
	sEntry, found := lookupNeighbour(t, objs.Neighbours, senderMAC)
	if !found {
		t.Fatal("sender MAC should be tracked from ARP reply")
	}
	if sEntry.Ipv4Count < 1 {
		t.Fatalf("expected sender ipv4_count>=1, got %d", sEntry.Ipv4Count)
	}
	if !containsIPv4(ipv4FromEntry(sEntry), senderIP) {
		t.Errorf("sender IP %s not found", senderIP)
	}

	// Target should also have its IP tracked
	tEntry, found := lookupNeighbour(t, objs.Neighbours, targetMAC)
	if !found {
		t.Fatal("target MAC should be tracked from ARP reply")
	}
	if tEntry.Ipv4Count < 1 {
		t.Fatalf("expected target ipv4_count>=1, got %d", tEntry.Ipv4Count)
	}
	if !containsIPv4(ipv4FromEntry(tEntry), targetIP) {
		t.Errorf("target IP %s not found", targetIP)
	}
}

func TestGratuitousARP(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	senderMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x30}
	ip := net.ParseIP("192.168.1.50").To4()
	broadcast := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	// Gratuitous ARP: sender IP == target IP
	pkt := buildARPPacket(broadcast, senderMAC, 1, senderMAC, ip, broadcast, ip)
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, senderMAC)
	if !found {
		t.Fatal("sender MAC should be tracked from gratuitous ARP")
	}
	if entry.Ipv4Count != 1 {
		t.Fatalf("expected ipv4_count=1, got %d", entry.Ipv4Count)
	}
	if !containsIPv4(ipv4FromEntry(entry), ip) {
		t.Errorf("gratuitous ARP IP %s not found", ip)
	}
}

func TestARPIPv4Dedup(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	senderMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x40}
	senderIP := net.ParseIP("192.168.1.100").To4()
	targetMAC := net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	targetIP := net.ParseIP("192.168.1.1").To4()
	broadcast := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	pkt := buildARPPacket(broadcast, senderMAC, 1, senderMAC, senderIP, targetMAC, targetIP)

	// Send same ARP twice
	runProgram(t, objs.L2radar, pkt)
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, senderMAC)
	if !found {
		t.Fatal("MAC not found")
	}
	if entry.Ipv4Count != 1 {
		t.Errorf("expected ipv4_count=1 after dedup, got %d", entry.Ipv4Count)
	}
}

func TestARPIPv4Cap(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	senderMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x50}
	targetMAC := net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	targetIP := net.ParseIP("192.168.1.1").To4()
	broadcast := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	// Send 5 ARP requests with different sender IPs
	for i := 0; i < 5; i++ {
		senderIP := net.IPv4(10, 0, 0, byte(i+1)).To4()
		pkt := buildARPPacket(broadcast, senderMAC, 1, senderMAC, senderIP, targetMAC, targetIP)
		runProgram(t, objs.L2radar, pkt)
	}

	entry, found := lookupNeighbour(t, objs.Neighbours, senderMAC)
	if !found {
		t.Fatal("MAC not found")
	}
	if entry.Ipv4Count != 4 {
		t.Errorf("expected ipv4_count=4 (capped), got %d", entry.Ipv4Count)
	}
}

// --- NDP Helpers ---

// buildIPv6Header constructs a minimal IPv6 header.
func buildIPv6Header(src, dst net.IP, nextHeader uint8, payloadLen uint16) []byte {
	hdr := make([]byte, 40)
	hdr[0] = 0x60 // version 6
	binary.BigEndian.PutUint16(hdr[4:6], payloadLen)
	hdr[6] = nextHeader
	hdr[7] = 255 // hop limit
	copy(hdr[8:24], src.To16())
	copy(hdr[24:40], dst.To16())
	return hdr
}

// buildNDPOption constructs a single NDP TLV option.
// optType: 1=Source Link-Layer Address, 2=Target Link-Layer Address
// length is in units of 8 bytes.
func buildNDPOption(optType uint8, mac net.HardwareAddr) []byte {
	opt := make([]byte, 8) // 1 unit = 8 bytes
	opt[0] = optType
	opt[1] = 1 // length in 8-byte units
	copy(opt[2:8], mac)
	return opt
}

// buildNDPNS constructs a Neighbor Solicitation ICMPv6 packet body.
// NS body: type(1) + code(1) + checksum(2) + reserved(4) + target(16) + options
func buildNDPNS(targetAddr net.IP, options []byte) []byte {
	body := make([]byte, 8+16+len(options))
	body[0] = 135 // ICMPv6 type: Neighbor Solicitation
	body[1] = 0   // code
	// checksum at [2:4] left as 0 (BPF doesn't validate)
	// reserved at [4:8]
	copy(body[8:24], targetAddr.To16())
	copy(body[24:], options)
	return body
}

// buildNDPNA constructs a Neighbor Advertisement ICMPv6 packet body.
// NA body: type(1) + code(1) + checksum(2) + flags(4) + target(16) + options
func buildNDPNA(targetAddr net.IP, solicited bool, options []byte) []byte {
	body := make([]byte, 8+16+len(options))
	body[0] = 136 // ICMPv6 type: Neighbor Advertisement
	body[1] = 0
	if solicited {
		body[4] = 0x40 // S flag (solicited)
	} else {
		body[4] = 0x20 // O flag (override, unsolicited)
	}
	copy(body[8:24], targetAddr.To16())
	copy(body[24:], options)
	return body
}

// buildNDPPacket constructs a full Ethernet+IPv6+NDP packet.
func buildNDPPacket(ethDst, ethSrc net.HardwareAddr, ipSrc, ipDst net.IP,
	ndpBody []byte) []byte {

	ipv6Hdr := buildIPv6Header(ipSrc, ipDst, 58, uint16(len(ndpBody)))
	payload := append(ipv6Hdr, ndpBody...)
	return buildEthernetFrame(ethDst, ethSrc, 0x86DD, payload)
}

// ipv6FromEntry extracts the active IPv6 addresses from a neighbour entry.
func ipv6FromEntry(entry *l2radarNeighbourEntry) []net.IP {
	var ips []net.IP
	for i := 0; i < int(entry.Ipv6Count); i++ {
		ip := make(net.IP, 16)
		copy(ip, entry.Ipv6[i].In6U.U6Addr8[:])
		ips = append(ips, ip)
	}
	return ips
}

// containsIPv6 checks if an IPv6 address is in the list.
func containsIPv6(ips []net.IP, target net.IP) bool {
	t := target.To16()
	for _, ip := range ips {
		if ip.Equal(t) {
			return true
		}
	}
	return false
}

// --- NDP Tests ---

func TestNDPNSWithSourceLinkLayerOption(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	srcMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x60}
	dstMAC := net.HardwareAddr{0x33, 0x33, 0xff, 0x11, 0x00, 0x01}
	srcIP := net.ParseIP("fe80::42:acff:fe11:60")
	dstIP := net.ParseIP("ff02::1:ff11:1")
	targetIP := net.ParseIP("fe80::42:acff:fe11:1")

	opts := buildNDPOption(1, srcMAC) // Source Link-Layer Address
	nsBody := buildNDPNS(targetIP, opts)
	pkt := buildNDPPacket(dstMAC, srcMAC, srcIP, dstIP, nsBody)
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, srcMAC)
	if !found {
		t.Fatal("source MAC should be tracked from NDP NS")
	}
	if entry.Ipv6Count < 1 {
		t.Fatalf("expected ipv6_count>=1, got %d", entry.Ipv6Count)
	}
	if !containsIPv6(ipv6FromEntry(entry), srcIP) {
		t.Errorf("source IPv6 %s not found in entry", srcIP)
	}
}

func TestNDPNAUnsolicited(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	srcMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x61}
	dstMAC := net.HardwareAddr{0x33, 0x33, 0x00, 0x00, 0x00, 0x01}
	srcIP := net.ParseIP("fe80::42:acff:fe11:61")
	dstIP := net.ParseIP("ff02::1")
	targetIP := net.ParseIP("2001:db8::1") // The address being announced

	opts := buildNDPOption(2, srcMAC) // Target Link-Layer Address
	naBody := buildNDPNA(targetIP, false, opts)
	pkt := buildNDPPacket(dstMAC, srcMAC, srcIP, dstIP, naBody)
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, srcMAC)
	if !found {
		t.Fatal("source MAC should be tracked from unsolicited NA")
	}
	// Should have both the IPv6 source and the NA target address
	if entry.Ipv6Count < 1 {
		t.Fatalf("expected ipv6_count>=1, got %d", entry.Ipv6Count)
	}
	ips := ipv6FromEntry(entry)
	if !containsIPv6(ips, targetIP) {
		t.Errorf("NA target IPv6 %s not found in entry", targetIP)
	}
}

func TestNDPNASolicited(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	srcMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x62}
	dstMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x63}
	srcIP := net.ParseIP("fe80::42:acff:fe11:62")
	dstIP := net.ParseIP("fe80::42:acff:fe11:63")
	targetIP := srcIP // NA target = source

	opts := buildNDPOption(2, srcMAC) // Target Link-Layer Address
	naBody := buildNDPNA(targetIP, true, opts)
	pkt := buildNDPPacket(dstMAC, srcMAC, srcIP, dstIP, naBody)
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, srcMAC)
	if !found {
		t.Fatal("source MAC should be tracked from solicited NA")
	}
	if entry.Ipv6Count < 1 {
		t.Fatalf("expected ipv6_count>=1, got %d", entry.Ipv6Count)
	}
	if !containsIPv6(ipv6FromEntry(entry), targetIP) {
		t.Errorf("NA target IPv6 %s not found", targetIP)
	}
}

func TestNDPIPv6Dedup(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	srcMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x64}
	dstMAC := net.HardwareAddr{0x33, 0x33, 0xff, 0x11, 0x00, 0x01}
	srcIP := net.ParseIP("fe80::42:acff:fe11:64")
	dstIP := net.ParseIP("ff02::1:ff11:1")
	targetIP := net.ParseIP("fe80::42:acff:fe11:1")

	opts := buildNDPOption(1, srcMAC)
	nsBody := buildNDPNS(targetIP, opts)
	pkt := buildNDPPacket(dstMAC, srcMAC, srcIP, dstIP, nsBody)

	// Send twice
	runProgram(t, objs.L2radar, pkt)
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, srcMAC)
	if !found {
		t.Fatal("MAC not found")
	}
	if entry.Ipv6Count != 1 {
		t.Errorf("expected ipv6_count=1 after dedup, got %d", entry.Ipv6Count)
	}
}

func TestNDPIPv6Cap(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	srcMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x65}
	dstMAC := net.HardwareAddr{0x33, 0x33, 0x00, 0x00, 0x00, 0x01}
	dstIP := net.ParseIP("ff02::1")

	// Send 5 unsolicited NAs with different target IPs
	for i := 0; i < 5; i++ {
		srcIP := net.ParseIP("fe80::42:acff:fe11:65")
		targetIP := net.ParseIP("2001:db8::1")
		targetIP[15] = byte(i + 1)

		opts := buildNDPOption(2, srcMAC)
		naBody := buildNDPNA(targetIP, false, opts)
		pkt := buildNDPPacket(dstMAC, srcMAC, srcIP, dstIP, naBody)
		runProgram(t, objs.L2radar, pkt)
	}

	entry, found := lookupNeighbour(t, objs.Neighbours, srcMAC)
	if !found {
		t.Fatal("MAC not found")
	}
	if entry.Ipv6Count != 4 {
		t.Errorf("expected ipv6_count=4 (capped), got %d", entry.Ipv6Count)
	}
}

// --- NDP RS/RA Helpers ---

// buildNDPRS constructs a Router Solicitation ICMPv6 packet body.
// RS body: type(1) + code(1) + checksum(2) + reserved(4) + options
func buildNDPRS(options []byte) []byte {
	body := make([]byte, 8+len(options))
	body[0] = 133 // ICMPv6 type: Router Solicitation
	body[1] = 0
	copy(body[8:], options)
	return body
}

// buildNDPRA constructs a Router Advertisement ICMPv6 packet body.
// RA body: type(1) + code(1) + checksum(2) + hop_limit(1) + flags(1) +
//          router_lifetime(2) + reachable_time(4) + retrans_timer(4) + options
func buildNDPRA(options []byte) []byte {
	body := make([]byte, 16+len(options))
	body[0] = 134 // ICMPv6 type: Router Advertisement
	body[1] = 0
	body[4] = 64 // current hop limit
	// flags, lifetime, reachable, retrans left as 0
	copy(body[16:], options)
	return body
}

// --- NDP RS/RA Tests ---

func TestNDPRSWithSourceLinkLayerOption(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	srcMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x70}
	dstMAC := net.HardwareAddr{0x33, 0x33, 0x00, 0x00, 0x00, 0x02}
	srcIP := net.ParseIP("fe80::42:acff:fe11:70")
	dstIP := net.ParseIP("ff02::2") // all-routers multicast

	opts := buildNDPOption(1, srcMAC) // Source Link-Layer Address
	rsBody := buildNDPRS(opts)
	pkt := buildNDPPacket(dstMAC, srcMAC, srcIP, dstIP, rsBody)
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, srcMAC)
	if !found {
		t.Fatal("source MAC should be tracked from NDP RS")
	}
	if entry.Ipv6Count < 1 {
		t.Fatalf("expected ipv6_count>=1, got %d", entry.Ipv6Count)
	}
	if !containsIPv6(ipv6FromEntry(entry), srcIP) {
		t.Errorf("source IPv6 %s not found in entry", srcIP)
	}
}

func TestNDPRAWithSourceLinkLayerOption(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	routerMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x71}
	dstMAC := net.HardwareAddr{0x33, 0x33, 0x00, 0x00, 0x00, 0x01}
	routerIP := net.ParseIP("fe80::42:acff:fe11:71")
	dstIP := net.ParseIP("ff02::1") // all-nodes multicast

	opts := buildNDPOption(1, routerMAC) // Source Link-Layer Address
	raBody := buildNDPRA(opts)
	pkt := buildNDPPacket(dstMAC, routerMAC, routerIP, dstIP, raBody)
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, routerMAC)
	if !found {
		t.Fatal("router MAC should be tracked from NDP RA")
	}
	if entry.Ipv6Count < 1 {
		t.Fatalf("expected ipv6_count>=1, got %d", entry.Ipv6Count)
	}
	if !containsIPv6(ipv6FromEntry(entry), routerIP) {
		t.Errorf("router IPv6 %s not found in entry", routerIP)
	}
}

// --- 802.1Q VLAN Tests ---

func TestVLANTaggedARPRequest(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	senderMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x80}
	senderIP := net.ParseIP("192.168.1.100").To4()
	targetMAC := net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	targetIP := net.ParseIP("192.168.1.1").To4()
	broadcast := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	// Build ARP payload
	arp := make([]byte, 28)
	binary.BigEndian.PutUint16(arp[0:2], 1)     // htype
	binary.BigEndian.PutUint16(arp[2:4], 0x0800) // ptype
	arp[4] = 6                                    // hlen
	arp[5] = 4                                    // plen
	binary.BigEndian.PutUint16(arp[6:8], 1)      // opcode: request
	copy(arp[8:14], senderMAC)
	copy(arp[14:18], senderIP)
	copy(arp[18:24], targetMAC)
	copy(arp[24:28], targetIP)

	pkt := buildVLANEthernetFrame(broadcast, senderMAC, 100, 0x0806, arp)
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, senderMAC)
	if !found {
		t.Fatal("sender MAC should be tracked from VLAN-tagged ARP")
	}
	if entry.Ipv4Count != 1 {
		t.Fatalf("expected ipv4_count=1, got %d", entry.Ipv4Count)
	}
	if !containsIPv4(ipv4FromEntry(entry), senderIP) {
		t.Errorf("sender IP %s not found", senderIP)
	}
}

func TestVLANTaggedNDPNS(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	srcMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x81}
	dstMAC := net.HardwareAddr{0x33, 0x33, 0xff, 0x11, 0x00, 0x01}
	srcIP := net.ParseIP("fe80::42:acff:fe11:81")
	dstIP := net.ParseIP("ff02::1:ff11:1")
	targetIP := net.ParseIP("fe80::42:acff:fe11:1")

	opts := buildNDPOption(1, srcMAC)
	nsBody := buildNDPNS(targetIP, opts)
	ipv6Hdr := buildIPv6Header(srcIP, dstIP, 58, uint16(len(nsBody)))
	payload := append(ipv6Hdr, nsBody...)

	pkt := buildVLANEthernetFrame(dstMAC, srcMAC, 200, 0x86DD, payload)
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, srcMAC)
	if !found {
		t.Fatal("source MAC should be tracked from VLAN-tagged NDP NS")
	}
	if entry.Ipv6Count < 1 {
		t.Fatalf("expected ipv6_count>=1, got %d", entry.Ipv6Count)
	}
	if !containsIPv6(ipv6FromEntry(entry), srcIP) {
		t.Errorf("source IPv6 %s not found", srcIP)
	}
}

func TestVLANTaggedUnicastMACTracked(t *testing.T) {
	objs, cleanup := loadTestObjects(t)
	defer cleanup()

	srcMAC := net.HardwareAddr{0x02, 0x42, 0xac, 0x11, 0x00, 0x82}
	dstMAC := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}

	// Non-ARP/NDP VLAN-tagged frame
	pkt := buildVLANEthernetFrame(dstMAC, srcMAC, 300, 0x0800, make([]byte, 46))
	runProgram(t, objs.L2radar, pkt)

	entry, found := lookupNeighbour(t, objs.Neighbours, srcMAC)
	if !found {
		t.Fatal("unicast MAC should be tracked from VLAN-tagged frame")
	}
	if entry.Ipv4Count != 0 {
		t.Errorf("expected ipv4_count=0, got %d", entry.Ipv4Count)
	}
}
