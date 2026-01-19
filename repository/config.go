package repository

import (
	"os"
	"path/filepath"
	"strings"
)

/*
The add config overrides the name and email in the configFile of the repository.
*/
func WriteConfig(gitDir, name, email string) error {
	configPath := filepath.Join(gitDir, configFile)

	// existing values (if any)
	var currName, currEmail string

	// read existing config if present
	if data, err := os.ReadFile(configPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)

			if after, ok := strings.CutPrefix(line, "name ="); ok {
				currName = strings.TrimSpace(after)
			}

			if after, ok := strings.CutPrefix(line, "email ="); ok {
				currEmail = strings.TrimSpace(after)
			}
		}
	}

	// override only provided values
	if name != "" {
		currName = name
	}
	if email != "" {
		currEmail = email
	}

	// build new config
	var buf strings.Builder
	buf.WriteString("[user]\n")

	if currName != "" {
		buf.WriteString("\tname = " + currName + "\n")
	}
	if currEmail != "" {
		buf.WriteString("\temail = " + currEmail + "\n")
	}

	return os.WriteFile(configPath, []byte(buf.String()), 0644)
}

/*
Read config returns the name and the email present in the config file.
*/
func ReadConfig(gitDir string) (name, email string) {
	configPath := filepath.Join(gitDir, configFile)

	if data, err := os.ReadFile(configPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)

			if after, ok := strings.CutPrefix(line, "name ="); ok {
				name = strings.TrimSpace(after)
			}

			if after, ok := strings.CutPrefix(line, "email ="); ok {
				email = strings.TrimSpace(after)
			}
		}
	}

	return
}
