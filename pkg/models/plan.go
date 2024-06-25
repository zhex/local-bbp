package models

type Plan struct {
	DefaultImage string    `yaml:"image"`
	Pipelines    *Pipeline `yaml:"pipelines"`
}

func NewPlan() *Plan {
	return &Plan{
		DefaultImage: "atlassian/default-image:latest",
	}
}
