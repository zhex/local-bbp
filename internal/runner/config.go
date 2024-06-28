package runner

type Config struct {
	HostProjectPath string
	WorkDir         string
	DefaultImage    string
	OutputDir       string
}

func NewConfig() *Config {
	return &Config{
		HostProjectPath: ".",
		DefaultImage:    "atlassian/default-image:4",
		WorkDir:         "/project",
		OutputDir:       "./bbp",
	}
}
