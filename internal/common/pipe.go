package common

import "strings"

func GetPipeImage(image string) string {
	ws := "bitbucketpipelines"
	parts := strings.Split(image, "/")
	if len(parts) == 1 {
		return ws + parts[0]
	}
	if len(parts) == 2 && parts[0] == "atlassian" {
		parts[0] = ws
		return strings.Join(parts, "/")
	}
	return image
}
