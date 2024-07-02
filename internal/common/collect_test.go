package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeMaps(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"b": 3, "c": 4}
	m3 := map[string]int{"d": 5}
	expected := map[string]int{"a": 1, "b": 3, "c": 4, "d": 5}
	result := MergeMaps(m1, m2, m3)
	assert.Equal(t, expected, result)
}

func TestContains(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	assert.True(t, Contains(slice, 1))
	assert.True(t, Contains(slice, 3))
	assert.False(t, Contains(slice, 6))
}
