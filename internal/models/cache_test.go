package models

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestCache_UnmarshalYAML(t *testing.T) {
	data := `
mytest: /tmp
`
	var c *Caches
	_ = yaml.Unmarshal([]byte(data), &c)
	cache := c.Get("mytest")
	assert.Equal(t, "/tmp", cache.Path)
	assert.Nil(t, cache.Key)

	data = `
mytest:
  path: /tmp
  key:
    files:
      - file1
      - file2
`
	var c2 *Caches
	_ = yaml.Unmarshal([]byte(data), &c2)
	cache = c2.Get("mytest")
	assert.Equal(t, "/tmp", cache.Path)
	assert.Equal(t, []string{"file1", "file2"}, cache.Key.Files)
}
