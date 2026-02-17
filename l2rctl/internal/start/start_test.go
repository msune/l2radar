package start

import (
	"fmt"
	"testing"

	"github.com/msune/l2radar/l2rctl/internal/docker"
)

func TestPullImage_RemoteSuccess(t *testing.T) {
	r := &docker.MockRunner{}
	if err := pullImage(r, "ghcr.io/msune/l2radar:latest"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Calls) != 1 || r.Calls[0][0] != "pull" {
		t.Fatalf("expected single pull call, got %v", r.Calls)
	}
}

func TestPullImage_PullFailsButLocalExists(t *testing.T) {
	r := &docker.MockRunner{
		ErrFn: func(args []string) error {
			if args[0] == "pull" {
				return fmt.Errorf("pull access denied")
			}
			return nil
		},
	}
	if err := pullImage(r, "my-local-image:latest"); err != nil {
		t.Fatalf("expected success (image exists locally), got: %v", err)
	}
	// Should have called pull, then image inspect
	if len(r.Calls) != 2 {
		t.Fatalf("expected 2 calls (pull + inspect), got %v", r.Calls)
	}
	if r.Calls[1][0] != "image" || r.Calls[1][1] != "inspect" {
		t.Fatalf("expected image inspect call, got %v", r.Calls[1])
	}
}

func TestPullImage_PullFailsAndLocalMissing(t *testing.T) {
	r := &docker.MockRunner{
		ErrFn: func(args []string) error {
			return fmt.Errorf("not found")
		},
	}
	err := pullImage(r, "nonexistent:latest")
	if err == nil {
		t.Fatal("expected error when pull fails and image not local")
	}
}

func TestParseTarget(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{"default is all", nil, "all", false},
		{"explicit all", []string{"all"}, "all", false},
		{"probe", []string{"probe"}, "probe", false},
		{"ui", []string{"ui"}, "ui", false},
		{"invalid", []string{"bogus"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTarget(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
