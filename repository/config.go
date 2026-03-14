package repository

import (
	"os"
	"path/filepath"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// Config
// ─────────────────────────────────────────────────────────────────────────────

// WriteConfig updates name and/or email in .gitingo/config.
// Blank arguments leave the existing value unchanged.
func WriteConfig(gitDir, name, email string) error {
	curr := ReadConfig(gitDir) // load whatever exists first

	if name != "" {
		curr.Name = name
	}
	if email != "" {
		curr.Email = email
	}

	var b strings.Builder
	b.WriteString("[user]\n")
	if curr.Name != "" {
		b.WriteString("\tname = " + curr.Name + "\n")
	}
	if curr.Email != "" {
		b.WriteString("\temail = " + curr.Email + "\n")
	}

	return os.WriteFile(filepath.Join(gitDir, configFile), []byte(b.String()), 0644)
}

// Config holds the user identity stored in .gitingo/config.
type Config struct {
	Name  string
	Email string
}

// ReadConfig parses .gitingo/config and returns the current user identity.
// Missing or unreadable config returns a zero Config (both fields empty).
func ReadConfig(gitDir string) Config {
	var cfg Config

	data, err := os.ReadFile(filepath.Join(gitDir, configFile))
	if err != nil {
		return cfg
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, "name ="); ok {
			cfg.Name = strings.TrimSpace(v)
		}
		if v, ok := strings.CutPrefix(line, "email ="); ok {
			cfg.Email = strings.TrimSpace(v)
		}
	}
	return cfg
}
