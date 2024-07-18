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

	needSave := false

	if c.DefaultDockerImage == "" {
		c.DefaultDockerImage = defaultConfig.DefaultDockerImage
		needSave = true
	}
	if c.MaxStepTimeout == 0 {
		c.MaxStepTimeout = defaultConfig.MaxStepTimeout
		needSave = true
	}
	if c.MaxPipelineTimeout == 0 {
		c.MaxPipelineTimeout = defaultConfig.MaxPipelineTimeout
		needSave = true
	}

	if needSave {
		// fix the config file for the missing field in new version
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
