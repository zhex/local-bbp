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

func GetGitChangedFiles(path string, branch string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	uncommittedChanges := strings.Split(trimOutput(out), "\n")

	if branch == "" {
		branch = "HEAD~1"
	}

	cmd2 := exec.Command("git", "diff", "--name-only", "HEAD", branch)
	cmd2.Dir = path
	out2, err := cmd2.Output()
	if err != nil {
		return nil, err
	}
	lastCommitChanges := strings.Split(trimOutput(out2), "\n")

	changes := make(map[string]bool)
	for _, file := range uncommittedChanges {
		changes[file] = true
	}
	for _, file := range lastCommitChanges {
		changes[file] = true
	}
	var getKeys = func(m map[string]bool) []string {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}
	return getKeys(changes), nil
}

func trimOutput(data []byte) string {
	return strings.Trim(string(data), "\r\n")
}
