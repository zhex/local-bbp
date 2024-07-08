package config

import (
	"encoding/json"
	"github.com/zhex/local-bbp/internal/common"
	"os"
	"path"
)

type Config struct {
	WorkDir           string `json:"workDir"`
	DefaultImage      string `json:"defaultImage"`
	OutputDir         string `json:"outputDir"`
	HostDockerDaemon  string `json:"hostDockerDaemon"`
	HostDockerCLIPath string `json:"hostDockerCLI"`
}

func NewConfig() *Config {
	home, _ := getConfigHome()

	return &Config{
		DefaultImage:      "atlassian/default-image:4",
		WorkDir:           "/opt/atlassian/pipelines/agent/build",
		OutputDir:         "bbp",
		HostDockerDaemon:  "/var/run/docker.sock",
		HostDockerCLIPath: path.Join(home, "tools/docker"),
	}
}

func (c *Config) Persistent() error {
	home, err := getConfigHome()
	if err != nil {
		return err
	}
	if !common.IsDirExists(home) {
		_ = os.Mkdir(home, 0755)
	}
	filePath := getConfigFile()
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
