package common

import gonanoid "github.com/matoous/go-nanoid/v2"

func NewID(prefix string) string {
	id, _ := gonanoid.Generate("1234567890abcdefghijklmnopqrstuvwxyz", 8)
	return prefix + id
}
