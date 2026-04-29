package agentdeck

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// configFileName is the name of the config file within the config directory.
const configFileName = "config.json"

// Config holds the user-facing configuration for agentdeck.
type Config struct {
}

// LoadConfig reads and decodes the config file from the given config directory.
// If the file does not exist, a zero-value Config is returned with no error;
// callers can treat a missing file as "use defaults".
func LoadConfig(configDir string) (*Config, error) {
	path := configPath(configDir)

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("open config %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("decode config %q: %w", path, err)
	}

	return &cfg, nil
}

// SaveConfig writes the config to disk atomically (write to temp, then rename).
func SaveConfig(configDir string, cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create config dir %q: %w", configDir, err)
	}

	path := configPath(configDir)

	tmp, err := os.CreateTemp(configDir, ".config-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpPath := tmp.Name()

	// Best-effort cleanup if anything goes wrong before the rename.
	cleanup := func() { _ = os.Remove(tmpPath) }

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("encode config: %w", err)
	}

	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp config: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return fmt.Errorf("rename temp config to %q: %w", path, err)
	}

	return nil
}

// configPath returns the full path to the config file within configDir.
func configPath(configDir string) string {
	return filepath.Join(configDir, configFileName)
}
