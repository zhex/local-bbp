package models

type StepTrigger string

const StepTriggerAutomatic StepTrigger = "automatic"
const StepTriggerManual StepTrigger = "manual"

type Step struct {
	Name        string            `yaml:"name"`
	Image       string            `yaml:"image"`
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
