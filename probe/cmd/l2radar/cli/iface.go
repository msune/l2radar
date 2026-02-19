package cli

import (
	"fmt"
	"net"
	"strings"
)

// listInterfaces is overridable for testing.
var listInterfaces = defaultListInterfaces

func defaultListInterfaces() ([]net.Interface, error) {
	return net.Interfaces()
}

// virtualPrefixes lists interface name prefixes considered virtual/infrastructure.
// These are filtered out by the "external" keyword.
var virtualPrefixes = []string{"veth", "docker", "br-", "virbr"}

// isVirtualInterface returns true if the interface name matches a virtual prefix.
func isVirtualInterface(name string) bool {
	for _, prefix := range virtualPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// resolveInterfaces expands "external" and "any" keywords to matching interfaces.
//   - "any": all L2 interfaces except loopbacks.
//   - "external": all external interfaces (excludes loopbacks and virtual interfaces
//     like docker*, veth*, br-*, virbr*).
//
// Non-keyword values are returned as-is. Duplicates are removed.
func resolveInterfaces(ifaces []string) ([]string, error) {
	seen := make(map[string]bool)
	var result []string
	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}

	for _, name := range ifaces {
		lower := strings.ToLower(name)
		if lower == "any" || lower == "external" {
			filterVirtual := lower == "external"
			allIfaces, err := listInterfaces()
			if err != nil {
				return nil, fmt.Errorf("listing interfaces: %w", err)
			}
			for _, iface := range allIfaces {
				if iface.Flags&net.FlagLoopback != 0 {
					continue
				}
				if filterVirtual && isVirtualInterface(iface.Name) {
					continue
				}
				add(iface.Name)
			}
		} else {
			add(name)
		}
	}
	return result, nil
}
