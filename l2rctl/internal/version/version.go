// Package version provides version checking for l2rctl against the Go module proxy.
package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

const (
	// Module is the Go module path used for go install.
	Module = "github.com/msune/l2radar/l2rctl"

	// ProxyURL is the default Go module proxy base URL.
	ProxyURL = "https://proxy.golang.org/" + Module + "/@latest"

	// cacheTTL is how long a cached version check remains valid.
	cacheTTL = 24 * time.Hour
)

// proxyResponse is the JSON returned by the Go module proxy /@latest endpoint.
type proxyResponse struct {
	Version string `json:"Version"`
}

// cacheEntry is persisted to disk between runs.
type cacheEntry struct {
	Latest    string    `json:"latest"`
	CheckedAt time.Time `json:"checked_at"`
}

// Current returns the version of the running binary as embedded by Go
// at build time. Returns "(devel)" for local/unversioned builds.
func Current() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	return info.Main.Version
}

// FetchLatest queries the Go module proxy at baseURL and returns the latest
// version string (e.g. "v1.2.3").
func FetchLatest(ctx context.Context, baseURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching latest version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("proxy returned status %d", resp.StatusCode)
	}

	var pr proxyResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return "", fmt.Errorf("decoding proxy response: %w", err)
	}
	return pr.Version, nil
}

// parseSemver extracts major, minor, patch from a "vX.Y.Z" string.
// Returns ok=false if the string is not valid semver.
func parseSemver(v string) (major, minor, patch int, ok bool) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	var err error
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, false
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, false
	}
	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, false
	}
	return major, minor, patch, true
}

// IsNewer reports whether latest is a newer semver than current.
// Returns false if either string is not valid semver.
func IsNewer(current, latest string) bool {
	cMaj, cMin, cPat, cOK := parseSemver(current)
	lMaj, lMin, lPat, lOK := parseSemver(latest)
	if !cOK || !lOK {
		return false
	}

	if lMaj != cMaj {
		return lMaj > cMaj
	}
	if lMin != cMin {
		return lMin > cMin
	}
	return lPat > cPat
}

// DefaultCacheFile returns the default path for the version cache file.
func DefaultCacheFile() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = filepath.Join(os.TempDir(), "l2rctl")
	} else {
		dir = filepath.Join(dir, "l2rctl")
	}
	return filepath.Join(dir, "version-check.json")
}

// CheckCached checks for a newer version, using a disk cache to avoid
// hitting the proxy on every invocation. Returns an upgrade hint message,
// or "" if the version is current (or on any error).
// Skips the check entirely for "(devel)" builds.
func CheckCached(ctx context.Context, current, cacheFile, proxyURL string) string {
	if current == "(devel)" || current == "" {
		return ""
	}

	// Try reading cache.
	if data, err := os.ReadFile(cacheFile); err == nil {
		var c cacheEntry
		if json.Unmarshal(data, &c) == nil && time.Since(c.CheckedAt) < cacheTTL {
			if IsNewer(current, c.Latest) {
				return UpgradeMsg(c.Latest)
			}
			return ""
		}
	}

	// Cache missing or expired — fetch.
	latest, err := FetchLatest(ctx, proxyURL)
	if err != nil {
		return ""
	}

	// Write cache (best-effort).
	c := cacheEntry{Latest: latest, CheckedAt: time.Now()}
	if data, err := json.Marshal(c); err == nil {
		os.MkdirAll(filepath.Dir(cacheFile), 0o755)
		os.WriteFile(cacheFile, data, 0o644)
	}

	if IsNewer(current, latest) {
		return UpgradeMsg(latest)
	}
	return ""
}

// UpgradeMsg returns a user-facing message about an available upgrade.
func UpgradeMsg(latest string) string {
	return fmt.Sprintf("A new version of l2rctl is available: %s\n"+
		"Update with: go install %s/cmd/l2rctl@latest", latest, Module)
}
