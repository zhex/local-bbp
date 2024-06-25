package models

type Action struct {
	Step     *Step     `yaml:"step"`
	Parallel *Parallel `yaml:"parallel"`
}

func (a *Action) IsParallel() bool {
	return a.Parallel != nil
}
