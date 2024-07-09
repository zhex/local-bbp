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

	data = `
image:
  name: xxx
  username: user
  password: pwd
`
	var v2 tmp
	_ = yaml.Unmarshal([]byte(data), &v2)
	assert.Equal(t, "xxx", v2.Image.Name)
	assert.Equal(t, "user", v2.Image.Username)
	assert.Equal(t, "pwd", v2.Image.Password)
}
