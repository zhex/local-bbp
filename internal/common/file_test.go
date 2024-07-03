package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetFileSha256(t *testing.T) {
	file := "testdata/file1.txt"
	sha256, _ := GetFileSha256(file)
	assert.Equal(t, "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2", sha256)

	file = "testdata/file_unknown.txt"
	sha256, err := GetFileSha256(file)
	assert.Error(t, err)
	assert.Equal(t, "", sha256)
}

func TestGetFilesSha256(t *testing.T) {
	files := []string{"testdata/file1.txt", "testdata/file2.txt"}
	sha256, _ := GetFilesSha256(files)
	assert.Equal(t, "c5790a940ee9cb2564f7c1077546b31b5e374dadd29b0389f2edf82ea4b81fb2", sha256)

	files = []string{"testdata/file1.txt", "testdata/file_unknown.txt"}
	sha256, err := GetFilesSha256(files)
	assert.Error(t, err)
	assert.Equal(t, "", sha256)
}
