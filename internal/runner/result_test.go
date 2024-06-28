package runner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStepResult_GetIdxString(t *testing.T) {
	sr := &StepResult{
		Index: 1,
	}
	assert.Equal(t, "1", sr.GetIdxString())

	sr.Index = 1.1
	assert.Equal(t, "1.1", sr.GetIdxString())

	sr.Index = 1.22
	assert.Equal(t, "1.22", sr.GetIdxString())
}
