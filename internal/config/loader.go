package config

import (
	"encoding/json"
	"github.com/zhex/local-bbp/internal/common"
	"os"
	"path/filepath"
)

func LoadConfig() (*Config, error) {
	filePath := GetConfigFile()
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

func GetConfigHome() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".bbp"), nil
}

func GetConfigFile() string {
	home, _ := GetConfigHome()
	return filepath.Join(home, "config.json")
}
