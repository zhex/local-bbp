package models

type Step struct {
	Name        string            `yaml:"name"`
	Image       string            `yaml:"image"`
	Script      string            `yaml:"script"`
	Environment map[string]string `yaml:"environment"`
}
