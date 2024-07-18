package models

type Action struct {
	Step     *Step     `yaml:"step"`
	Parallel *Parallel `yaml:"parallel"`
	Stage    *Stage    `yaml:"stage"`
}

func (a *Action) IsStep() bool {
	return a.Step != nil
}

func (a *Action) IsStage() bool {
	return a.Stage != nil
}

func (a *Action) IsParallel() bool {
	return a.Parallel != nil
}
