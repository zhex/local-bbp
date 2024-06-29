package models

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type Artifact struct {
	Paths    []string
	Download bool
}

func (a *Artifact) UnmarshalYAML(node *yaml.Node) error {
	var stringList []string
	if err := node.Decode(&stringList); err == nil {
		a.Paths = stringList
		a.Download = true
		return nil
	}

	var object struct {
		Paths    []string `yaml:"paths"`
		Download bool     `yaml:"download"`
	}
	if err := node.Decode(&object); err == nil {
		a.Paths = object.Paths
		a.Download = object.Download
		return nil
	}

	return fmt.Errorf("failed to unmarshal artifacts")
}
