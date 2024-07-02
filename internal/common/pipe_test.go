package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPipeImage(t *testing.T) {
	image := "atlassian/default-image"
	expected := "bitbucketpipelines/default-image"
	result := GetPipeImage(image)
	assert.Equal(t, expected, result)

	image = "bitbucketpipelines/default-image"
	expected = "bitbucketpipelines/default-image"
	result = GetPipeImage(image)
	assert.Equal(t, expected, result)

	image = "default-image"
	expected = "default-image"
	result = GetPipeImage(image)
	assert.Equal(t, expected, result)
}
