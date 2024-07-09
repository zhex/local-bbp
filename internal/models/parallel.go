package models

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type Parallel struct {
	Actions  []*Action `yaml:"steps"`
	FailFast bool      `yaml:"fail-fast"`
}

func (p *Parallel) UnmarshalYAML(value *yaml.Node) error {
	var tmp struct {
		Actions  []*Action `yaml:"steps"`
		FailFast bool      `yaml:"fail-fast"`
	}

	switch value.Kind {
	case yaml.MappingNode:
		if err := value.Decode(&tmp); err != nil {
			return err
		}
		p.Actions = tmp.Actions
		p.FailFast = tmp.FailFast
	case yaml.SequenceNode:
		if err := value.Decode(&p.Actions); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected YAML node kind: %v", value.Kind)
	}
	return nil
}
