package models

import "gopkg.in/yaml.v3"

type Image struct {
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (i *Image) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		return value.Decode(&i.Name)
	}

	var tmp struct {
		Name     string `yaml:"name"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}

	if err := value.Decode(&tmp); err != nil {
		return err
	}

	i.Name = tmp.Name
	i.Username = tmp.Username
	i.Password = tmp.Password

	return nil
}
