package runner

type Config struct {
	WorkDir      string
	DefaultImage string
	OutputDir    string
}

func NewConfig() *Config {
	return &Config{
		DefaultImage: "atlassian/default-image:4",
		WorkDir:      "/project",
		OutputDir:    "./bbp",
	}
}
