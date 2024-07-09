package models

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestParallel_UnmarshalYAML(t *testing.T) {
	data := `
steps:
- step: 
   name: test1
- step: 
   name: test2
fail-fast: true
`
	var o Parallel
	_ = yaml.Unmarshal([]byte(data), &o)
	assert.Equal(t, 2, len(o.Actions))
	assert.Equal(t, true, o.FailFast)

	data = `
- step: 
   name: test1
- step: 
   name: test2
`
	var o2 Parallel
	_ = yaml.Unmarshal([]byte(data), &o2)
	assert.Equal(t, 2, len(o2.Actions))
	assert.Equal(t, false, o2.FailFast)
}
