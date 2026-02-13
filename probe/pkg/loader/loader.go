package loader

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

const (
	// DefaultPinPath is the default base path for pinning eBPF maps.
	DefaultPinPath = "/sys/fs/bpf/l2radar"

	// MapPinPermissions is the file mode for pinned maps (world-readable).
	MapPinPermissions os.FileMode = 0444
)

// Probe represents an attached eBPF probe on a network interface.
type Probe struct {
	iface   string
	objs    *l2radarObjects
	link    link.Link
	pinPath string
	logger  *slog.Logger
}

// Attach loads the eBPF program, attaches it to the given interface via
// TCX ingress, and pins the neighbours map at <pinBase>/neigh-<iface>.
func Attach(iface string, pinBase string, logger *slog.Logger) (*Probe, error) {
	if logger == nil {
		logger = slog.Default()
	}

	if err := rlimit.RemoveMemlock(); err != nil {
		logger.Warn("failed to remove memlock rlimit", "error", err)
	}

	ifObj, err := net.InterfaceByName(iface)
	if err != nil {
		return nil, fmt.Errorf("interface %s: %w", iface, err)
	}

	var objs l2radarObjects
	if err := loadL2radarObjects(&objs, &ebpf.CollectionOptions{}); err != nil {
		return nil, fmt.Errorf("loading eBPF objects: %w", err)
	}

	// Pin the map
	mapPinPath := filepath.Join(pinBase, fmt.Sprintf("neigh-%s", iface))
	if err := os.MkdirAll(pinBase, 0755); err != nil {
		objs.Close()
		return nil, fmt.Errorf("creating pin directory %s: %w", pinBase, err)
	}

	if err := objs.Neighbours.Pin(mapPinPath); err != nil {
		objs.Close()
		return nil, fmt.Errorf("pinning map at %s: %w", mapPinPath, err)
	}

	// Set world-readable permissions on the pinned map
	if err := os.Chmod(mapPinPath, MapPinPermissions); err != nil {
		os.Remove(mapPinPath)
		objs.Close()
		return nil, fmt.Errorf("setting map permissions: %w", err)
	}

	// Attach via TCX ingress
	tcxLink, err := link.AttachTCX(link.TCXOptions{
		Interface: ifObj.Index,
		Program:   objs.L2radar,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		os.Remove(mapPinPath)
		objs.Close()
		return nil, fmt.Errorf("attaching TCX to %s: %w", iface, err)
	}

	logger.Info("probe attached",
		"interface", iface,
		"ifindex", ifObj.Index,
		"pin_path", mapPinPath,
	)

	return &Probe{
		iface:   iface,
		objs:    &objs,
		link:    tcxLink,
		pinPath: mapPinPath,
		logger:  logger,
	}, nil
}

// Close detaches the eBPF program and unpins the map.
func (p *Probe) Close() error {
	var errs []error

	if p.link != nil {
		if err := p.link.Close(); err != nil {
			errs = append(errs, fmt.Errorf("detaching link: %w", err))
		}
	}

	if p.pinPath != "" {
		if err := os.Remove(p.pinPath); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("removing pin %s: %w", p.pinPath, err))
		}
	}

	if p.objs != nil {
		if err := p.objs.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing objects: %w", err))
		}
	}

	p.logger.Info("probe detached", "interface", p.iface)

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}

// Interface returns the name of the interface this probe is attached to.
func (p *Probe) Interface() string {
	return p.iface
}

// MapPinPath returns the filesystem path where the map is pinned.
func (p *Probe) MapPinPath() string {
	return p.pinPath
}
