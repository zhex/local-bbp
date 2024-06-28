package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewID(t *testing.T) {
	id := NewID("test-")
	assert.Equal(t, 15, len(id))
	assert.Equal(t, "test-", id[:5])
}
