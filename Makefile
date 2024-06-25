VERSION ?= $(shell git describe --tags --always --dirty)

.PHONY: build
build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/bbp main.go
