package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the user's obx configuration
type Config struct {
	ActiveVault string            `json:"active_vault"`
	Vaults      map[string]string `json:"vaults"`
}

// GetConfigPath returns the path to the obx configuration file
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not get user config directory: %v", err)
	}

	obxDir := filepath.Join(configDir, "obx")
	return filepath.Join(obxDir, "config.json"), nil
}

// Load reads the configuration from the file system
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Vaults: make(map[string]string),
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default empty config if file doesn't exist
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	if cfg.Vaults == nil {
		cfg.Vaults = make(map[string]string)
	}

	return cfg, nil
}

// Save writes the configuration to the file system
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	obxDir := filepath.Dir(configPath)
	if err := os.MkdirAll(obxDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	return nil
}

// GetActiveVaultPath returns the filesystem path of the currently active vault
func (c *Config) GetActiveVaultPath() (string, error) {
	if c.ActiveVault == "" {
		return "", fmt.Errorf("no active vault configured")
	}

	path, ok := c.Vaults[c.ActiveVault]
	if !ok {
		return "", fmt.Errorf("active vault '%s' not found in known vaults", c.ActiveVault)
	}

	return path, nil
}
