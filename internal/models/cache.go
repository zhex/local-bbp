package models

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type CacheKey struct {
	Files []string `yaml:"files"`
}

type Cache struct {
	Path string    `yaml:"path"`
	Key  *CacheKey `yaml:"key"`
}

func NewCache(path string, files []string) *Cache {
	return &Cache{
		Path: path,
		Key: &CacheKey{
			Files: files,
		},
	}
}

func (c *Cache) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		var path string
		if err := value.Decode(&path); err != nil {
			return err
		}
		c.Path = path
	} else if value.Kind == yaml.MappingNode {
		type cache Cache
		if err := value.Decode((*cache)(c)); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown cache type: %v", value.Kind)
	}
	return nil
}

type Caches map[string]*Cache

func (c Caches) Get(name string) *Cache {
	if cache, ok := c[name]; ok {
		return cache
	}
	return nil
}

func (c Caches) Set(name string, cache *Cache) {
	c[name] = cache
}
