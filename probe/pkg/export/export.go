package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/marc/l2radar/probe/pkg/dump"
)

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
	Interface  string          `json:"interface"`
	Timestamp  string          `json:"timestamp"`
	Neighbours []NeighbourJSON `json:"neighbours"`
}

// NewInterfaceData converts dump.Neighbour entries to the JSON export format.
func NewInterfaceData(iface string, ts time.Time, neighbours []dump.Neighbour) InterfaceData {
	data := InterfaceData{
		Interface:  iface,
		Timestamp:  ts.UTC().Format(time.RFC3339),
		Neighbours: make([]NeighbourJSON, 0, len(neighbours)),
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
// so readers never see a partial file.
func WriteJSON(iface string, neighbours []dump.Neighbour, outputDir string, ts time.Time) error {
	data := NewInterfaceData(iface, ts, neighbours)

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

	if err := os.Rename(tmpName, outPath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}
