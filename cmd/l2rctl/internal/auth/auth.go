package auth

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type userEntry struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type authConfig struct {
	Users []userEntry `yaml:"users"`
}

// ParseUser splits "user:pass" on the first colon.
func ParseUser(s string) (string, string, error) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return "", "", fmt.Errorf("invalid user format %q: missing ':'", s)
	}
	user := s[:idx]
	pass := s[idx+1:]
	if user == "" {
		return "", "", fmt.Errorf("invalid user format %q: empty username", s)
	}
	if pass == "" {
		return "", "", fmt.Errorf("invalid user format %q: empty password", s)
	}
	return user, pass, nil
}

// GenerateAuthYAML produces auth.yaml content from user:pass strings.
func GenerateAuthYAML(users []string) (string, error) {
	if len(users) == 0 {
		return "", fmt.Errorf("no users specified")
	}
	var entries []userEntry
	for _, u := range users {
		user, pass, err := ParseUser(u)
		if err != nil {
			return "", err
		}
		entries = append(entries, userEntry{Username: user, Password: pass})
	}
	data, err := yaml.Marshal(authConfig{Users: entries})
	if err != nil {
		return "", fmt.Errorf("marshal auth yaml: %w", err)
	}
	return string(data), nil
}

// WriteAuthFile writes auth.yaml to a temp file and returns the path.
func WriteAuthFile(users []string) (string, error) {
	content, err := GenerateAuthYAML(users)
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp("", "l2rctl-auth-*.yaml")
	if err != nil {
		return "", fmt.Errorf("create temp auth file: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("write auth file: %w", err)
	}
	return f.Name(), nil
}

// ValidateFlags checks that --user-file and --user are mutually exclusive.
func ValidateFlags(userFile string, users []string) error {
	if userFile != "" && len(users) > 0 {
		return fmt.Errorf("--user-file and --user are mutually exclusive")
	}
	return nil
}
