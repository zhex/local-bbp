package models

type Parallel struct {
	Actions  []*Action `yaml:"steps"`
	FailFast bool      `yaml:"fail-fast"`
}
