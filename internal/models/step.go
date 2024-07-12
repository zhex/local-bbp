package models

import "github.com/bmatcuk/doublestar/v4"

type StepTrigger string

const StepTriggerAutomatic StepTrigger = "automatic"
const StepTriggerManual StepTrigger = "manual"

type Step struct {
	Name        string            `yaml:"name"`
	Image       *Image            `yaml:"image"`
	Script      StepScript        `yaml:"script"`
	AfterScript []string          `yaml:"after-script"`
	Environment map[string]string `yaml:"environment"`
	MaxTime     int               `yaml:"max-time"`
	Size        string            `yaml:"size"`
	Deployment  string            `yaml:"deployment"`
	Trigger     StepTrigger       `yaml:"trigger"`
	Artifacts   *Artifact         `yaml:"artifacts"`
	Caches      []string          `yaml:"caches"`
	Services    []string          `yaml:"services"`
	RunsOn      []string          `yaml:"runs-on"`
	Condition   *Condition        `yaml:"condition"`
}

func (s *Step) IsManual() bool {
	return s.Trigger == StepTriggerManual
}

func (s *Step) GetName() string {
	if s.Name == "" {
		return "default"
	}
	return s.Name
}

func (s *Step) HasImage() bool {
	return s.Image != nil && s.Image.Name != ""
}

func (s *Step) MatchCondition(changedFiles []string) bool {
	if s.Condition == nil {
		return true
	}
	for _, pattern := range s.Condition.ChangeSets.IncludePaths {
		for _, changedFile := range changedFiles {
			matched, err := doublestar.PathMatch(pattern, changedFile)
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}
