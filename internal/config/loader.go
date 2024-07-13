package config

import (
	"encoding/json"
	"github.com/zhex/local-bbp/internal/common"
	"os"
	"path/filepath"
)

func LoadConfig() (*Config, error) {
	defaultConfig := NewConfig()

	filePath := GetConfigFile()
	if !common.IsFileExists(filePath) {
		c := defaultConfig
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

	if c.DefaultDockerImage == "" {
		// fix the config file for the missing field in new version
		c.DefaultDockerImage = defaultConfig.DefaultDockerImage
		err := c.Persistent()
		if err != nil {
			return nil, err
		}
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
