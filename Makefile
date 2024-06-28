.PHONY: build
build:
	go build -o bin/bbp main.go

.PHONY: test
test:
	go test -v ./...
