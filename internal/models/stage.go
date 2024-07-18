package models

import "github.com/bmatcuk/doublestar/v4"

type Stage struct {
	Name       string       `yaml:"name"`
	Actions    []*Action    `yaml:"steps"`
	Condition  *Condition   `yaml:"condition"`
	Trigger    *StepTrigger `yaml:"trigger"`
	Deployment string       `yaml:"deployment"`
}

func (s *Stage) MatchCondition(changedFiles []string) bool {
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
