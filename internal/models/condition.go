package models

type ChangeSet struct {
	IncludePaths []string `yaml:"includePaths"`
}

type Condition struct {
	ChangeSets ChangeSet `yaml:"changesets"`
}
