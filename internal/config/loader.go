package config

import (
	"encoding/json"
	"github.com/zhex/local-bbp/internal/common"
	"os"
	"path/filepath"
)

func LoadConfig() (*Config, error) {
	filePath := getConfigFile()
	if !common.IsFileExists(filePath) {
		c := NewConfig()
		err := c.Persistent()
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	f, err := os.Open(filePath)
	var c Config
	err = json.NewDecoder(f).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func getConfigHome() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".bbp"), nil
}

func getConfigFile() string {
	home, _ := getConfigHome()
	return filepath.Join(home, "config.json")
}
