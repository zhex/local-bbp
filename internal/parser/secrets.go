package parser

import (
	"regexp"
	"strings"
)

func ParseSecrets(data []byte) map[string]string {
	secrets := make(map[string]string)
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		trimPattern, _ := regexp.Compile(`^[\s'"]+|[\s'"]+$`)
		value := trimPattern.ReplaceAllString(parts[1], "")
		secrets[key] = value
	}

	return secrets
}
