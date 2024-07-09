package models

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"gopkg.in/yaml.v3"
	"strings"
)

type AWSAuth struct {
	AccessKey string `yaml:"access-key"`
	SecretKey string `yaml:"secret-key"`
	OIDCRole  string `yaml:"oidc-role"`
}

func extractAwsRegionFromImage(image string) string {
	parts := strings.Split(image, ".")
	if len(parts) < 4 {
		panic("Invalid image name format")
	}
	return parts[3]
}

func (a *AWSAuth) GetAuthData(image string) (*ecr.AuthorizationData, error) {
	region := extractAwsRegionFromImage(image)
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(a.AccessKey, a.SecretKey, ""),
	})
	if err != nil {
		return nil, err
	}
	svc := ecr.New(sess)
	token, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return nil, err
	}
	return token.AuthorizationData[0], nil
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
