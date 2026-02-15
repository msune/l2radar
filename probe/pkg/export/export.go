package export

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/marc/l2radar/probe/pkg/dump"
)

// InterfaceInfo holds the monitored interface's own addresses.
type InterfaceInfo struct {
	MAC  net.HardwareAddr
	IPv4 []net.IP
	IPv6 []net.IP
}

// LookupInterfaceInfo returns the MAC and IP addresses of a network interface.
func LookupInterfaceInfo(name string) (*InterfaceInfo, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("looking up interface %s: %w", name, err)
	}

	info := &InterfaceInfo{
		MAC:  iface.HardwareAddr,
		IPv4: []net.IP{},
		IPv6: []net.IP{},
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("listing addresses for %s: %w", name, err)
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if v4 := ipNet.IP.To4(); v4 != nil {
			info.IPv4 = append(info.IPv4, v4)
		} else {
			info.IPv6 = append(info.IPv6, ipNet.IP)
		}
	}

	return info, nil
}

// InterfaceStats holds kernel TX/RX counters for a network interface.
type InterfaceStats struct {
	TxBytes   uint64 `json:"tx_bytes"`
	RxBytes   uint64 `json:"rx_bytes"`
	TxPackets uint64 `json:"tx_packets"`
	RxPackets uint64 `json:"rx_packets"`
	TxErrors  uint64 `json:"tx_errors"`
	RxErrors  uint64 `json:"rx_errors"`
	TxDropped uint64 `json:"tx_dropped"`
	RxDropped uint64 `json:"rx_dropped"`
}

// readSysStat reads a single counter from /sys/class/net/<iface>/statistics/<name>.
func readSysStat(iface, name string) (uint64, error) {
	path := fmt.Sprintf("/sys/class/net/%s/statistics/%s", iface, name)
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
}

// LookupInterfaceStats reads kernel interface counters from sysfs.
func LookupInterfaceStats(name string) (*InterfaceStats, error) {
	fields := []string{
		"tx_bytes", "rx_bytes",
		"tx_packets", "rx_packets",
		"tx_errors", "rx_errors",
		"tx_dropped", "rx_dropped",
	}
	vals := make([]uint64, len(fields))
	for i, f := range fields {
		v, err := readSysStat(name, f)
		if err != nil {
			return nil, fmt.Errorf("reading %s for %s: %w", f, name, err)
		}
		vals[i] = v
	}
	return &InterfaceStats{
		TxBytes:   vals[0],
		RxBytes:   vals[1],
		TxPackets: vals[2],
		RxPackets: vals[3],
		TxErrors:  vals[4],
		RxErrors:  vals[5],
		TxDropped: vals[6],
		RxDropped: vals[7],
	}, nil
}

// NeighbourJSON is the JSON representation of a neighbour entry.
type NeighbourJSON struct {
	MAC       string   `json:"mac"`
	IPv4      []string `json:"ipv4"`
	IPv6      []string `json:"ipv6"`
	FirstSeen string   `json:"first_seen"`
	LastSeen  string   `json:"last_seen"`
}

// InterfaceData is the top-level JSON structure for one interface export.
type InterfaceData struct {
	Interface      string          `json:"interface"`
	Timestamp      string          `json:"timestamp"`
	ExportInterval string          `json:"export_interval"`
	MAC            string          `json:"mac"`
	IPv4           []string        `json:"ipv4"`
	IPv6           []string        `json:"ipv6"`
	Stats          *InterfaceStats `json:"stats"`
	Neighbours     []NeighbourJSON `json:"neighbours"`
}

// NewInterfaceData converts dump.Neighbour entries to the JSON export format.
// ifInfo and stats may be nil if unavailable.
func NewInterfaceData(iface string, ts time.Time, interval time.Duration, neighbours []dump.Neighbour, ifInfo *InterfaceInfo, stats *InterfaceStats) InterfaceData {
	data := InterfaceData{
		Interface:      iface,
		Timestamp:      ts.UTC().Format(time.RFC3339),
		ExportInterval: interval.String(),
		IPv4:           []string{},
		IPv6:           []string{},
		Stats:          stats,
		Neighbours:     make([]NeighbourJSON, 0, len(neighbours)),
	}

	if ifInfo != nil {
		data.MAC = ifInfo.MAC.String()
		for _, ip := range ifInfo.IPv4 {
			data.IPv4 = append(data.IPv4, ip.String())
		}
		for _, ip := range ifInfo.IPv6 {
			data.IPv6 = append(data.IPv6, ip.String())
		}
	}

	for _, n := range neighbours {
		nj := NeighbourJSON{
			MAC:       n.MAC.String(),
			IPv4:      make([]string, 0, len(n.IPv4)),
			IPv6:      make([]string, 0, len(n.IPv6)),
			FirstSeen: n.FirstSeen.UTC().Format(time.RFC3339),
			LastSeen:  n.LastSeen.UTC().Format(time.RFC3339),
		}
		for _, ip := range n.IPv4 {
			nj.IPv4 = append(nj.IPv4, ip.String())
		}
		for _, ip := range n.IPv6 {
			nj.IPv6 = append(nj.IPv6, ip.String())
		}
		data.Neighbours = append(data.Neighbours, nj)
	}

	return data
}

// OutputFileName returns the JSON file name for an interface.
func OutputFileName(iface string) string {
	return fmt.Sprintf("neigh-%s.json", iface)
}

// WriteJSON writes the neighbour data for an interface to a JSON file
// in the given output directory. The write is atomic (temp file + rename)
// so readers never see a partial file. ifInfo may be nil.
func WriteJSON(iface string, neighbours []dump.Neighbour, outputDir string, ts time.Time, interval time.Duration, ifInfo *InterfaceInfo, stats *InterfaceStats) error {
	data := NewInterfaceData(iface, ts, interval, neighbours, ifInfo, stats)

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	b = append(b, '\n')

	outPath := filepath.Join(outputDir, OutputFileName(iface))

	// Write to temp file in the same directory, then rename for atomicity.
	tmp, err := os.CreateTemp(outputDir, ".neigh-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}

	// Make world-readable so nginx (unprivileged) can serve the file.
	if err := os.Chmod(tmpName, 0644); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("setting file permissions: %w", err)
	}

	if err := os.Rename(tmpName, outPath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}
