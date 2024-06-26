package models

type Pipe struct {
	Name      string            `yaml:"name"`
	Image     string            `yaml:"pipe"`
	Variables map[string]string `yaml:"variables"`
}
