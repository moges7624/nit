package repo

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	UserName  string
	UserEmail string
}

func (r *Repository) LoadConfig() (*Config, error) {
	cfg := &Config{}

	// try local .git/config
	localConfig := filepath.Join(r.nitPath, "config")
	if data, err := os.ReadFile(localConfig); err == nil {
		parseConfig(data, cfg)
	}

	// try global config (~/.gitconfig)
	if cfg.UserName == "" || cfg.UserEmail == "" {
		if home, err := os.UserHomeDir(); err == nil {
			globalConfig := filepath.Join(home, ".gitconfig")
			if data, err := os.ReadFile(globalConfig); err == nil {
				parseConfig(data, cfg)
			}
		}
	}

	return cfg, nil
}

func parseConfig(data []byte, cfg *Config) {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	inUserSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "[user]") {
			inUserSection = true
			continue
		}
		if strings.HasPrefix(line, "[") && inUserSection {
			inUserSection = false
		}

		if !inUserSection || line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if idx := strings.Index(line, "="); idx != -1 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			switch key {
			case "name":
				cfg.UserName = value
			case "email":
				cfg.UserEmail = value
			}
		}
	}
}
