package main

import (
	"testing"
)

func TestParseSubcommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{"start", []string{"l2rctl", "start"}, "start", false},
		{"stop", []string{"l2rctl", "stop"}, "stop", false},
		{"status", []string{"l2rctl", "status"}, "status", false},
		{"dump", []string{"l2rctl", "dump"}, "dump", false},
		{"no args", []string{"l2rctl"}, "", true},
		{"unknown", []string{"l2rctl", "bogus"}, "", true},
		{"help flag", []string{"l2rctl", "--help"}, "help", false},
		{"help short", []string{"l2rctl", "-h"}, "help", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSubcommand(tt.args)
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
