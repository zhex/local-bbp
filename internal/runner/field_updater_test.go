package runner

import (
	"github.com/stretchr/testify/assert"
	"github.com/zhex/local-bbp/internal/models"
	"testing"
)

func TestFieldUpdater_UpdateImage(t *testing.T) {
	secrets := map[string]string{
		"IMAGE_PASSWD": "pass",
		"IMAGE_USER":   "user",
	}

	image := &models.Image{
		Name:     "alpine",
		Username: "$IMAGE_USER",
		Password: "$IMAGE_PASSWD",
	}

	newImage := NewFieldUpdater(secrets).UpdateImage(image)
	assert.NotEqual(t, *image, *newImage)
	assert.Equal(t, "alpine", newImage.Name)
	assert.Equal(t, "user", newImage.Username)
	assert.Equal(t, "pass", newImage.Password)

}

func TestFieldUpdater_UpdateMap(t *testing.T) {
	secrets := map[string]string{
		"AWS_ACCESS_KEY": "access_key",
		"AWS_SECRET":     "secret",
	}

	data := map[string]string{
		"key1": "$AWS_ACCESS_KEY",
		"key2": "$AWS_SECRET",
		"key3": "xx_${AWS_SECRET}_xx",
		"key4": "xx_${AWS_SECRET1}_xx",
	}

	newData := NewFieldUpdater(secrets).UpdateMap(data)
	assert.Equal(t, "access_key", newData["key1"])
	assert.Equal(t, "secret", newData["key2"])
	assert.Equal(t, "xx_secret_xx", newData["key3"])
	assert.Equal(t, "xx__xx", newData["key4"])
}
