package oui

import (
	_ "embed"
	"fmt"
	"net"
	"strings"
)

//go:embed oui.txt
var ouiData string

// db maps uppercase OUI prefix (e.g. "28-6F-B9") to vendor name.
var db map[string]string

func init() {
	db = parse(ouiData)
}

// parse extracts OUI entries from the IEEE oui.txt format.
// Lines matching "XX-XX-XX   (hex)\t\tVendor Name" are extracted.
func parse(data string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(data, "\n") {
		// Look for lines containing "(hex)" — these have the format:
		// "28-6F-B9   (hex)\t\tNokia Shanghai Bell Co., Ltd."
		idx := strings.Index(line, "(hex)")
		if idx < 0 {
			continue
		}
		prefix := strings.TrimSpace(line[:idx])
		if len(prefix) != 8 { // "XX-XX-XX"
			continue
		}
		vendor := strings.TrimSpace(line[idx+len("(hex)"):])
		if vendor != "" {
			m[strings.ToUpper(prefix)] = vendor
		}
	}
	return m
}

// Lookup returns the vendor name for a MAC address OUI prefix.
// Returns empty string if the MAC is nil, too short, or not found.
func Lookup(mac net.HardwareAddr) string {
	if len(mac) < 3 {
		return ""
	}
	key := fmt.Sprintf("%02X-%02X-%02X", mac[0], mac[1], mac[2])
	return db[key]
}

// EntryCount returns the number of OUI entries in the database.
func EntryCount() int {
	return len(db)
}
