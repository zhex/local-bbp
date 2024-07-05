package models

type Plan struct {
	DefaultImage string      `yaml:"image"`
	Pipelines    *Pipeline   `yaml:"pipelines"`
	Definitions  *Definition `yaml:"definitions"`
}
