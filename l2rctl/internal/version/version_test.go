package version

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCurrent(t *testing.T) {
	v := Current()
	if v == "" {
		t.Fatal("Current() returned empty string")
	}
	// In tests, Go embeds "(devel)" since there's no module version.
	if v != "(devel)" {
		t.Logf("Current() = %q (expected '(devel)' in test)", v)
	}
}

func TestFetchLatest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := proxyResponse{Version: "v1.2.3"}
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		v, err := FetchLatest(context.Background(), srv.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "v1.2.3" {
			t.Fatalf("got %q, want %q", v, "v1.2.3")
		}
	})

	t.Run("server error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		_, err := FetchLatest(context.Background(), srv.URL)
		if err == nil {
			t.Fatal("expected error for 500 response")
		}
	})

	t.Run("bad json", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer srv.Close()

		_, err := FetchLatest(context.Background(), srv.URL)
		if err == nil {
			t.Fatal("expected error for bad JSON")
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Second)
		}))
		defer srv.Close()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := FetchLatest(ctx, srv.URL)
		if err == nil {
			t.Fatal("expected error for cancelled context")
		}
	})
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"v1.0.0", "v1.0.1", true},
		{"v1.0.0", "v1.1.0", true},
		{"v1.0.0", "v2.0.0", true},
		{"v1.0.1", "v1.0.0", false},
		{"v1.0.0", "v1.0.0", false},
		{"v0.1.0", "v0.2.0", true},
		{"v0.9.9", "v1.0.0", true},
		{"v1.2.3", "v1.2.3", false},
		// Non-semver current (e.g. "(devel)") should return false.
		{"(devel)", "v1.0.0", false},
		// Non-semver latest should return false.
		{"v1.0.0", "bad", false},
		// Both non-semver.
		{"(devel)", "bad", false},
	}
	for _, tt := range tests {
		t.Run(tt.current+"_vs_"+tt.latest, func(t *testing.T) {
			got := IsNewer(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("IsNewer(%q, %q) = %v, want %v",
					tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestCheckCached(t *testing.T) {
	t.Run("fresh cache no upgrade", func(t *testing.T) {
		dir := t.TempDir()
		cacheFile := filepath.Join(dir, "version-check.json")

		c := &cacheEntry{
			Latest:    "v1.0.0",
			CheckedAt: time.Now(),
		}
		data, _ := json.Marshal(c)
		os.WriteFile(cacheFile, data, 0o644)

		msg := CheckCached(context.Background(), "v1.0.0", cacheFile, "http://unused")
		if msg != "" {
			t.Fatalf("expected empty message, got %q", msg)
		}
	})

	t.Run("fresh cache with upgrade", func(t *testing.T) {
		dir := t.TempDir()
		cacheFile := filepath.Join(dir, "version-check.json")

		c := &cacheEntry{
			Latest:    "v2.0.0",
			CheckedAt: time.Now(),
		}
		data, _ := json.Marshal(c)
		os.WriteFile(cacheFile, data, 0o644)

		msg := CheckCached(context.Background(), "v1.0.0", cacheFile, "http://unused")
		if msg == "" {
			t.Fatal("expected upgrade message")
		}
	})

	t.Run("expired cache fetches new", func(t *testing.T) {
		dir := t.TempDir()
		cacheFile := filepath.Join(dir, "version-check.json")

		// Write expired cache.
		c := &cacheEntry{
			Latest:    "v1.0.0",
			CheckedAt: time.Now().Add(-48 * time.Hour),
		}
		data, _ := json.Marshal(c)
		os.WriteFile(cacheFile, data, 0o644)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := proxyResponse{Version: "v2.0.0"}
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		msg := CheckCached(context.Background(), "v1.0.0", cacheFile, srv.URL)
		if msg == "" {
			t.Fatal("expected upgrade message after cache refresh")
		}
	})

	t.Run("no cache fetches new", func(t *testing.T) {
		dir := t.TempDir()
		cacheFile := filepath.Join(dir, "version-check.json")

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := proxyResponse{Version: "v1.0.0"}
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		msg := CheckCached(context.Background(), "v1.0.0", cacheFile, srv.URL)
		if msg != "" {
			t.Fatalf("expected empty message (same version), got %q", msg)
		}

		// Verify cache was written.
		if _, err := os.Stat(cacheFile); err != nil {
			t.Fatalf("cache file not created: %v", err)
		}
	})

	t.Run("devel version skips check", func(t *testing.T) {
		msg := CheckCached(context.Background(), "(devel)", "/nonexistent", "http://unused")
		if msg != "" {
			t.Fatalf("expected empty message for (devel), got %q", msg)
		}
	})
}
