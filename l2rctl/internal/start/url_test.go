package start

import (
	"testing"
)

func TestBuildAccessURLsHTTPSOnly(t *testing.T) {
	urls := BuildAccessURLs(12443, 12080, "127.0.0.1", false, "admin", "secret")
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(urls))
	}
	want := "https://admin:secret@localhost:12443"
	if urls[0] != want {
		t.Errorf("got %q, want %q", urls[0], want)
	}
}

func TestBuildAccessURLsHTTPSAndHTTP(t *testing.T) {
	urls := BuildAccessURLs(12443, 12080, "127.0.0.1", true, "admin", "secret")
	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urls))
	}
	wantHTTPS := "https://admin:secret@localhost:12443"
	wantHTTP := "http://admin:secret@localhost:12080"
	if urls[0] != wantHTTPS {
		t.Errorf("got %q, want %q", urls[0], wantHTTPS)
	}
	if urls[1] != wantHTTP {
		t.Errorf("got %q, want %q", urls[1], wantHTTP)
	}
}

func TestBuildAccessURLsNoCreds(t *testing.T) {
	urls := BuildAccessURLs(12443, 12080, "127.0.0.1", false, "", "")
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(urls))
	}
	want := "https://localhost:12443"
	if urls[0] != want {
		t.Errorf("got %q, want %q", urls[0], want)
	}
}

func TestBuildAccessURLsBindAllInterfaces(t *testing.T) {
	urls := BuildAccessURLs(12443, 12080, "0.0.0.0", false, "u", "p")
	want := "https://u:p@localhost:12443"
	if urls[0] != want {
		t.Errorf("got %q, want %q", urls[0], want)
	}
}

func TestBuildAccessURLsCustomBind(t *testing.T) {
	urls := BuildAccessURLs(443, 80, "192.168.1.10", true, "u", "p")
	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urls))
	}
	wantHTTPS := "https://u:p@192.168.1.10:443"
	wantHTTP := "http://u:p@192.168.1.10:80"
	if urls[0] != wantHTTPS {
		t.Errorf("got %q, want %q", urls[0], wantHTTPS)
	}
	if urls[1] != wantHTTP {
		t.Errorf("got %q, want %q", urls[1], wantHTTP)
	}
}
