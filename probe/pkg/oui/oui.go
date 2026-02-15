package oui

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
)

//go:embed oui.json
var ouiData []byte

// db maps lowercase colon-separated OUI prefix (e.g. "28:6f:b9") to vendor name.
var db map[string]string

func init() {
	if err := json.Unmarshal(ouiData, &db); err != nil {
		panic(fmt.Sprintf("oui: failed to parse embedded oui.json: %v", err))
	}
}

// Lookup returns the vendor name for a MAC address OUI prefix.
// Returns empty string if the MAC is nil, too short, or not found.
func Lookup(mac net.HardwareAddr) string {
	if len(mac) < 3 {
		return ""
	}
	key := fmt.Sprintf("%02x:%02x:%02x", mac[0], mac[1], mac[2])
	return db[key]
}

// EntryCount returns the number of OUI entries in the database.
func EntryCount() int {
	return len(db)
}
