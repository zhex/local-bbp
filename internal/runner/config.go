package runner

type Config struct {
	WorkDir      string
	DefaultImage string
}

func NewConfig() *Config {
	return &Config{
		DefaultImage: "atlassian/default-image:latest",
		WorkDir:      "/project",
	}
}
