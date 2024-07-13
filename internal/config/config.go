package config

import (
	"encoding/json"
	"github.com/zhex/local-bbp/internal/common"
	"os"
	"path"
)

type Config struct {
	WorkDir            string `json:"workDir"`
	DefaultImage       string `json:"defaultImage"`
	OutputDir          string `json:"outputDir"`
	DockerVersion      string `json:"dockerVersion"`
	DefaultDockerImage string `json:"defaultDockerImage"`
	ToolDir            string `json:"toolDir"`
}

func NewConfig() *Config {
	home, _ := GetConfigHome()

	return &Config{
		DefaultImage:       "atlassian/default-image:4",
		WorkDir:            "/opt/atlassian/pipelines/agent/build",
		OutputDir:          "bbp",
		DockerVersion:      "19.03.15",
		DefaultDockerImage: "docker:27.0.3-dind-alpine3.20",
		ToolDir:            path.Join(home, "tools"),
	}
}

func (c *Config) Persistent() error {
	home, err := GetConfigHome()
	if err != nil {
		return err
	}
	if !common.IsDirExists(home) {
		_ = os.Mkdir(home, 0755)
	}
	filePath := GetConfigFile()
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}
