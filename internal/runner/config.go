package runner

import "path/filepath"

type Config struct {
	HostProjectPath  string
	WorkDir          string
	DefaultImage     string
	OutputDir        string
	HostDockerDaemon string
	HostDockerCLI    string
}

func NewConfig() *Config {
	output, _ := filepath.Abs("./bbp")

	return &Config{
		HostProjectPath:  ".",
		DefaultImage:     "atlassian/default-image:4",
		WorkDir:          "/opt/atlassian/pipelines/agent/build",
		OutputDir:        output,
		HostDockerDaemon: "/var/run/docker.sock",
		HostDockerCLI:    "/Users/zhex/Downloads/docker/docker",
	}
}
