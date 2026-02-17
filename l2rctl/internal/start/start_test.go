package start

import (
	"testing"
)

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
