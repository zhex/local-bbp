package common

import (
	"os/exec"
	"strings"
)

func GetGitCommit(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return trimOutput(out), nil
}

func GetGitBranch(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return trimOutput(out), nil
}

func GetGitOwner(path string) (string, error) {
	cmd := exec.Command("git", "log", "--reverse", "--pretty=format:%an", "-n", "1")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return trimOutput(out), nil
}

func trimOutput(data []byte) string {
	return strings.Trim(string(data), "\r\n")
}
