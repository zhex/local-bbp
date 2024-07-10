package models

type Service struct {
	Image     *Image            `yaml:"image"`
	Memory    int               `yaml:"memory"`
	Variables map[string]string `yaml:"variables"`
	Type      string            `yaml:"type"`
}
