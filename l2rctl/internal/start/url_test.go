package start

import (
	"testing"
)

func TestBuildAccessURLsHTTPSOnly(t *testing.T) {
	urls := BuildAccessURLs(12443, 12080, "127.0.0.1", false)
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(urls))
	}
	want := "https://localhost:12443"
	if urls[0] != want {
		t.Errorf("got %q, want %q", urls[0], want)
	}
}

func TestBuildAccessURLsHTTPSAndHTTP(t *testing.T) {
	urls := BuildAccessURLs(12443, 12080, "127.0.0.1", true)
	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urls))
	}
	wantHTTPS := "https://localhost:12443"
	wantHTTP := "http://localhost:12080"
	if urls[0] != wantHTTPS {
		t.Errorf("got %q, want %q", urls[0], wantHTTPS)
	}
	if urls[1] != wantHTTP {
		t.Errorf("got %q, want %q", urls[1], wantHTTP)
	}
}

func TestBuildAccessURLsBindAllInterfaces(t *testing.T) {
	urls := BuildAccessURLs(12443, 12080, "0.0.0.0", false)
	want := "https://localhost:12443"
	if urls[0] != want {
		t.Errorf("got %q, want %q", urls[0], want)
	}
}

func TestBuildAccessURLsCustomBind(t *testing.T) {
	urls := BuildAccessURLs(443, 80, "192.168.1.10", true)
	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urls))
	}
	wantHTTPS := "https://192.168.1.10:443"
	wantHTTP := "http://192.168.1.10:80"
	if urls[0] != wantHTTPS {
		t.Errorf("got %q, want %q", urls[0], wantHTTPS)
	}
	if urls[1] != wantHTTP {
		t.Errorf("got %q, want %q", urls[1], wantHTTP)
	}
}
