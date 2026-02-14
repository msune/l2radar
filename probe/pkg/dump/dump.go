package dump

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"golang.org/x/sys/unix"

	"github.com/cilium/ebpf"
)

// MacKey mirrors the eBPF mac_key struct layout.
type MacKey struct {
	Addr [6]uint8
	Pad  [2]uint8
}

// In6Addr mirrors struct in6_addr layout from the BPF map.
type In6Addr struct {
	Bytes [16]uint8
}

// NeighbourEntry mirrors the eBPF neighbour_entry struct layout.
type NeighbourEntry struct {
	Ipv4      [4]uint32
	Ipv6      [4]In6Addr
	Ipv4Count uint8
	Ipv6Count uint8
	Pad       [6]uint8
	FirstSeen uint64
	LastSeen  uint64
}

// Neighbour is the user-facing representation of a neighbour entry.
type Neighbour struct {
	MAC       net.HardwareAddr
	IPv4      []net.IP
	IPv6      []net.IP
	FirstSeen time.Time
	LastSeen  time.Time
}

// IPv4String returns IPv4 addresses as a comma-separated string.
func (n *Neighbour) IPv4String() string {
	strs := make([]string, len(n.IPv4))
	for i, ip := range n.IPv4 {
		strs[i] = ip.String()
	}
	return strings.Join(strs, ", ")
}

// IPv6String returns IPv6 addresses as a comma-separated string.
func (n *Neighbour) IPv6String() string {
	strs := make([]string, len(n.IPv6))
	for i, ip := range n.IPv6 {
		strs[i] = ip.String()
	}
	return strings.Join(strs, ", ")
}

// PinPath returns the expected map pin path for an interface.
func PinPath(pinBase, iface string) string {
	return filepath.Join(pinBase, fmt.Sprintf("neigh-%s", iface))
}

// timeNow and monoNow are overridable for testing.
var (
	timeNow = time.Now
	monoNow = defaultMonoNow
)

// defaultMonoNow returns nanoseconds since boot via CLOCK_BOOTTIME.
// This matches bpf_ktime_get_boot_ns() used in the BPF program, which
// includes time spent suspended (unlike CLOCK_MONOTONIC).
func defaultMonoNow() int64 {
	var ts unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_BOOTTIME, &ts); err != nil {
		return 0
	}
	return int64(ts.Sec)*1e9 + int64(ts.Nsec)
}

// ktimeToTime converts a bpf_ktime_get_boot_ns value to wall-clock time.
// ktime is CLOCK_BOOTTIME nanoseconds since boot (including suspend).
// We derive the boot instant by subtracting CLOCK_BOOTTIME from wall clock.
func ktimeToTime(ktime uint64) time.Time {
	if ktime == 0 {
		return time.Time{}
	}
	now := timeNow()
	mono := monoNow()
	bootTime := now.Add(-time.Duration(mono))
	return bootTime.Add(time.Duration(ktime))
}

// ReadMap opens a pinned BPF map and reads all neighbour entries.
func ReadMap(pinPath string) ([]Neighbour, error) {
	m, err := ebpf.LoadPinnedMap(pinPath, nil)
	if err != nil {
		return nil, fmt.Errorf("opening pinned map %s: %w", pinPath, err)
	}
	defer m.Close()

	var (
		key    MacKey
		val    NeighbourEntry
		result []Neighbour
	)

	iter := m.Iterate()
	for iter.Next(&key, &val) {
		n := entryToNeighbour(key, val)
		result = append(result, n)
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("iterating map: %w", err)
	}

	return result, nil
}

// entryToNeighbour converts raw map key/value to a Neighbour.
func entryToNeighbour(key MacKey, val NeighbourEntry) Neighbour {
	n := Neighbour{
		MAC:       net.HardwareAddr(key.Addr[:]),
		FirstSeen: ktimeToTime(val.FirstSeen),
		LastSeen:  ktimeToTime(val.LastSeen),
	}

	for i := 0; i < int(val.Ipv4Count) && i < 4; i++ {
		ip := make(net.IP, 4)
		binary.LittleEndian.PutUint32(ip, val.Ipv4[i])
		n.IPv4 = append(n.IPv4, ip)
	}

	for i := 0; i < int(val.Ipv6Count) && i < 4; i++ {
		ip := make(net.IP, 16)
		copy(ip, val.Ipv6[i].Bytes[:])
		n.IPv6 = append(n.IPv6, ip)
	}

	return n
}

// SortByLastSeen sorts neighbours by LastSeen descending (most recent first).
func SortByLastSeen(neighbours []Neighbour) {
	sort.Slice(neighbours, func(i, j int) bool {
		return neighbours[i].LastSeen.After(neighbours[j].LastSeen)
	})
}

// FormatTable writes a formatted table of neighbours to the writer.
func FormatTable(w io.Writer, neighbours []Neighbour) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "MAC\tIPv4\tIPv6\tFIRST SEEN\tLAST SEEN")
	fmt.Fprintln(tw, "---\t----\t----\t----------\t---------")

	for _, n := range neighbours {
		firstSeen := ""
		lastSeen := ""
		if !n.FirstSeen.IsZero() {
			firstSeen = n.FirstSeen.Format("2006-01-02 15:04:05")
		}
		if !n.LastSeen.IsZero() {
			lastSeen = n.LastSeen.Format("2006-01-02 15:04:05")
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			n.MAC.String(),
			n.IPv4String(),
			n.IPv6String(),
			firstSeen,
			lastSeen,
		)
	}

	tw.Flush()
}
