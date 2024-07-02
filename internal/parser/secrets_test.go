package parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseSecrets(t *testing.T) {
	data := []byte(`# comment
key1="value1"
key2='value2'
`)
	secrets := ParseSecrets(data)
	assert.Equal(t, 2, len(secrets))
	assert.Equal(t, "value1", secrets["key1"])
	assert.Equal(t, "value2", secrets["key2"])
}
