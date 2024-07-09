package models

import "gopkg.in/yaml.v3"

type AWSAuth struct {
	AccessKey string `yaml:"access-key"`
	SecretKey string `yaml:"secret-key"`
	OIDCRole  string `yaml:"oidc-role"`
}

type Image struct {
	Name      string   `yaml:"name"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	RunAsUser int      `yaml:"run-as-user"`
	AWS       *AWSAuth `yaml:"aws"`
}

func (i *Image) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		return value.Decode(&i.Name)
	}

	var tmp struct {
		Name      string   `yaml:"name"`
		Username  string   `yaml:"username"`
		Password  string   `yaml:"password"`
		RunAsUser int      `yaml:"run-as-user"`
		AWS       *AWSAuth `yaml:"aws"`
	}

	if err := value.Decode(&tmp); err != nil {
		return err
	}

	i.Name = tmp.Name
	i.Username = tmp.Username
	i.Password = tmp.Password
	i.RunAsUser = tmp.RunAsUser
	i.AWS = tmp.AWS

	return nil
}
