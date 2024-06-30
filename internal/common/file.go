package common

import (
	"os"
	"os/exec"
)

func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func Untar(path, target string) error {
	return exec.Command("tar", "--no-same-owner", "-xvf", path, "-C", target).Run()
}
