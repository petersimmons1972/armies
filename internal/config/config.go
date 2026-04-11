// Package config handles loading and validation of ~/.armies/config.yaml.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds user configuration from ~/.armies/config.yaml.
type Config struct {
	RemoteURL    string `yaml:"remote_url"`
	DefaultModel string `yaml:"default_model"`
	ProfilesDir  string `yaml:"profiles_dir"`
}

// defaults returns a Config populated with default values for the given home directory.
func defaults(home string) Config {
	return Config{
		RemoteURL:    "",
		DefaultModel: "sonnet",
		ProfilesDir:  filepath.Join(home, ".armies", "profiles"),
	}
}

// LoadFrom reads config from path, merging over defaults.
// A missing file is not an error — returns pure defaults.
func LoadFrom(path string) (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("resolve home directory: %w", err)
	}

	d := defaults(home)
	cfg := d

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("read config %q: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %q: %w", path, err)
	}

	// Re-apply defaults for fields left at zero value by the YAML.
	// This prevents an explicit `profiles_dir: ""` in config from
	// silently clobbering the default with an empty string.
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = d.DefaultModel
	}
	if cfg.ProfilesDir == "" {
		cfg.ProfilesDir = d.ProfilesDir
	}

	return cfg, nil
}

// Load reads from the default config location (~/.armies/config.yaml).
func Load() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("resolve home directory: %w", err)
	}
	return LoadFrom(filepath.Join(home, ".armies", "config.yaml"))
}

// ProfilesDirLexicallyValidated resolves the profiles directory and validates
// that its lexical path stays within the user's home directory. This is a
// syntactic check only — it does NOT resolve symlinks, so a symlink inside
// home pointing outside home will still pass. Full symlink resolution is not
// performed because the profile directory may not exist yet at validation time.
// Returns an error if ProfilesDir is empty or if the resolved path escapes home.
func (c Config) ProfilesDirLexicallyValidated() (string, error) {
	if strings.TrimSpace(c.ProfilesDir) == "" {
		return "", fmt.Errorf("profiles_dir is not set; check ~/.armies/config.yaml")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	// Expand $HOME and bare ~ manually; os.ExpandEnv does not handle ~.
	expanded := os.ExpandEnv(c.ProfilesDir)
	if len(expanded) > 0 && expanded[0] == '~' {
		expanded = home + expanded[1:]
	}

	resolved, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("resolve profiles_dir %q: %w", expanded, err)
	}
	homeResolved, err := filepath.Abs(home)
	if err != nil {
		return "", fmt.Errorf("resolve home directory %q: %w", home, err)
	}

	rel, err := filepath.Rel(homeResolved, resolved)
	if err != nil || (len(rel) >= 2 && rel[:2] == "..") {
		return "", fmt.Errorf(
			"profiles_dir %q is outside your home directory %q; check ~/.armies/config.yaml",
			resolved, homeResolved,
		)
	}
	return resolved, nil
}

// ArmiesDir returns the path to the ~/.armies directory.
// Returns an error if the home directory cannot be determined.
func ArmiesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".armies"), nil
}

// MalusLedgerPath returns the path to the malus ledger YAML file.
// Returns an error if the home directory cannot be determined.
func MalusLedgerPath() (string, error) {
	dir, err := ArmiesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "accountability", "malus-ledger.yaml"), nil
}
