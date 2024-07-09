package models

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestImage_UnmarshalYAML(t *testing.T) {
	type tmp struct {
		Image Image `yaml:"image"`
	}

	data := `
image: xxx
`
	var v tmp
	_ = yaml.Unmarshal([]byte(data), &v)
	assert.Equal(t, "xxx", v.Image.Name)
	assert.Equal(t, "", v.Image.Username)
	assert.Equal(t, "", v.Image.Password)
	assert.Equal(t, 0, v.Image.RunAsUser)

	data = `
image:
  name: xxx
  username: user
  password: pwd
  run-as-user: 1000
`
	var v2 tmp
	_ = yaml.Unmarshal([]byte(data), &v2)
	assert.Equal(t, "xxx", v2.Image.Name)
	assert.Equal(t, "user", v2.Image.Username)
	assert.Equal(t, "pwd", v2.Image.Password)
	assert.Equal(t, 1000, v2.Image.RunAsUser)
}
