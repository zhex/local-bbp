package models

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type ScriptType string

const ScriptTypeCmd ScriptType = "cmd"
const ScriptTypePipe ScriptType = "pipe"

type ScriptItem interface {
	Type() ScriptType
}

type CmdScript struct {
	Cmd string
}

func (c *CmdScript) Type() ScriptType {
	return ScriptTypeCmd
}

type Pipe struct {
	Name      string            `yaml:"name"`
	Pipe      string            `yaml:"pipe"`
	Variables map[string]string `yaml:"variables"`
}

func (p *Pipe) Type() ScriptType {
	return ScriptTypePipe
}

type StepScript []ScriptItem

func (s *StepScript) UnmarshalYAML(value *yaml.Node) error {
	var items []yaml.Node
	if err := value.Decode(&items); err != nil {
		return err
	}

	for _, item := range items {
		if item.Kind == yaml.ScalarNode {
			var cmd string
			if err := item.Decode(&cmd); err != nil {
				return err
			}
			cmdScript := &CmdScript{Cmd: cmd}
			*s = append(*s, cmdScript)
		} else if item.Kind == yaml.MappingNode {
			var pipe Pipe
			if err := item.Decode(&pipe); err != nil {
				return err
			}
			*s = append(*s, &pipe)
		} else {
			return fmt.Errorf("unknown script step type: %v", item.Kind)
		}
	}
	return nil
}

func (s *StepScript) HasPipe() bool {
	for _, item := range *s {
		if item.Type() == ScriptTypePipe {
			return true
		}
	}
	return false
}
