package models

type Definition struct {
	Caches   Caches              `yaml:"caches"`
	Services map[string]*Service `yaml:"services"`
}
