package models

import "strings"

type Plan struct {
	DefaultImage *Image      `yaml:"image"`
	Pipelines    *Pipeline   `yaml:"pipelines"`
	Definitions  *Definition `yaml:"definitions"`
}

func (p *Plan) GetPipeline(name string) []*Action {
	if name == "default" || name == "" {
		return p.Pipelines.Default
	}
	if strings.HasPrefix(name, "branch/") {
		name = strings.TrimPrefix(name, "branch/")
		return p.Pipelines.Branches[name]
	}
	if strings.HasPrefix(name, "pr/") {
		name = strings.TrimPrefix(name, "pr/")
		return p.Pipelines.PullRequests[name]
	}
	if strings.HasPrefix(name, "tag/") {
		name = strings.TrimPrefix(name, "tag/")
		return p.Pipelines.Tags[name]
	}
	if _, ok := p.Pipelines.Custom[name]; ok {
		return p.Pipelines.Custom[name]
	}
	return nil
}

func (p *Plan) GetCaches() Caches {
	if p.Definitions == nil {
		return nil
	}
	return p.Definitions.Caches
}

func (p *Plan) HasImage() bool {
	return p.DefaultImage != nil && p.DefaultImage.Name != ""
}
