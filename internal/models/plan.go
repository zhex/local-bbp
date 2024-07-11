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
	if strings.HasPrefix(name, "custom/") {
		name = strings.TrimPrefix(name, "custom/")
		return p.Pipelines.Custom[name]
	}
	return nil
}

func (p *Plan) GetPipelineNames() []string {
	names := make([]string, 0)
	if p.Pipelines.Default != nil {
		names = append(names, "default")
	}
	for name := range p.Pipelines.Custom {
		names = append(names, "custom/"+name)
	}
	for name := range p.Pipelines.Branches {
		names = append(names, "branch/"+name)
	}
	for name := range p.Pipelines.PullRequests {
		names = append(names, "pr/"+name)
	}
	for name := range p.Pipelines.Tags {
		names = append(names, "tag/"+name)
	}
	return names

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
