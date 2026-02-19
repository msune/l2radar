package auth

import (
	"os"
	"strings"
	"testing"
)

func TestParseUser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantUser string
		wantPass string
		wantErr  bool
	}{
		{"valid", "admin:secret", "admin", "secret", false},
		{"colon in password", "admin:sec:ret", "admin", "sec:ret", false},
		{"no colon", "admin", "", "", true},
		{"empty username", ":secret", "", "", true},
		{"empty password", "admin:", "", "", true},
		{"empty both", ":", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, p, err := ParseUser(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u != tt.wantUser {
				t.Errorf("username = %q, want %q", u, tt.wantUser)
			}
			if p != tt.wantPass {
				t.Errorf("password = %q, want %q", p, tt.wantPass)
			}
		})
	}
}

func TestGenerateAuthYAML(t *testing.T) {
	t.Run("single user", func(t *testing.T) {
		got, err := GenerateAuthYAML([]string{"admin:secret"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "users:\n    - username: admin\n      password: secret\n"
		if got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("multiple users", func(t *testing.T) {
		got, err := GenerateAuthYAML([]string{"admin:secret", "bob:pass123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "admin") || !strings.Contains(got, "bob") {
			t.Errorf("missing users in output:\n%s", got)
		}
	})

	t.Run("invalid user", func(t *testing.T) {
		_, err := GenerateAuthYAML([]string{"badformat"})
		if err == nil {
			t.Fatal("expected error for invalid user format")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		_, err := GenerateAuthYAML([]string{})
		if err == nil {
			t.Fatal("expected error for empty list")
		}
	})
}

func TestWriteAuthFile(t *testing.T) {
	path, err := WriteAuthFile([]string{"admin:secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read file: %v", err)
	}
	if !strings.Contains(string(data), "admin") {
		t.Errorf("file content missing 'admin': %s", data)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("cannot stat file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("file permissions = %o, want 600", got)
	}
}

func TestGenerateRandomCredentials(t *testing.T) {
	t.Run("format", func(t *testing.T) {
		cred, err := GenerateRandomCredentials()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		user, pass, err := ParseUser(cred)
		if err != nil {
			t.Fatalf("credential not parseable: %v", err)
		}
		if !strings.HasPrefix(user, "admin") {
			t.Errorf("username %q does not start with 'admin'", user)
		}
		if len(user) != 9 { // "admin" (5) + 4 hex chars
			t.Errorf("username length = %d, want 9", len(user))
		}
		if len(pass) != 16 {
			t.Errorf("password length = %d, want 16", len(pass))
		}
	})

	t.Run("unique", func(t *testing.T) {
		a, _ := GenerateRandomCredentials()
		b, _ := GenerateRandomCredentials()
		if a == b {
			t.Errorf("two calls returned identical credentials: %q", a)
		}
	})
}

func TestValidateFlags(t *testing.T) {
	t.Run("both set", func(t *testing.T) {
		err := ValidateFlags("file.yaml", []string{"admin:pass"})
		if err == nil {
			t.Fatal("expected error when both flags set")
		}
	})

	t.Run("neither set", func(t *testing.T) {
		err := ValidateFlags("", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("only file", func(t *testing.T) {
		err := ValidateFlags("file.yaml", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("only users", func(t *testing.T) {
		err := ValidateFlags("", []string{"admin:pass"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
