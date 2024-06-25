package models

type Pipeline struct {
	Default      []*Action            `yaml:"default"`
	Branches     map[string][]*Action `yaml:"branches"`
	PullRequests map[string][]*Action `yaml:"pull-requests"`
	Tags         map[string][]*Action `yaml:"tags"`
	Custom       map[string][]*Action `yaml:"custom"`
}
