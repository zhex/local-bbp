package common

import (
	"strconv"
	"time"
)

func NewID(prefix string) string {
	id := strconv.FormatInt(time.Now().Unix(), 10)
	return prefix + id
}
