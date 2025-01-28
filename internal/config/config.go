package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const cfgFileaname = ".gatorconfig.json"

type Config struct {
	Db_url       string
	Current_user string
}

func (c *Config) SetUser(currentUser string) error {
	c.Current_user = currentUser
	homePath, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting path to home directory: %w", err)
	}
	cfgData, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err = os.WriteFile(filepath.Join(homePath, cfgFileaname), cfgData, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	return nil
}

func Read() (*Config, error) {
	var cfgReturn Config
	homePath, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting path to home directory: %w", err)
	}
	configFile, err := os.ReadFile(filepath.Join(homePath, cfgFileaname))
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := json.Unmarshal(configFile, &cfgReturn); err != nil {
		return nil, fmt.Errorf("error unmarshaling config file: %w", err)
	}

	return &cfgReturn, nil
}
