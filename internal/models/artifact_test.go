package models

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

type obj struct {
	Artifacts *Artifact `yaml:"artifacts"`
}

func TestArtifact_UnmarshalYAML(t *testing.T) {
	data := `
artifacts:
  - file1.txt
  - file2.txt
`
	var o obj
	_ = yaml.Unmarshal([]byte(data), &o)
	assert.Equal(t, 2, len(o.Artifacts.Paths))
	assert.Equal(t, false, o.Artifacts.Download)

	data2 := `
artifacts:
  paths:
    - file1.txt
    - file2.txt
  download: true
`
	var o2 obj
	_ = yaml.Unmarshal([]byte(data2), &o2)
	assert.Equal(t, 2, len(o2.Artifacts.Paths))
	assert.Equal(t, true, o2.Artifacts.Download)
}
