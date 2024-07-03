#!/usr/bin/env bash

go install github.com/zhex/local-bbp@latest
mv $(go env GOPATH)/bin/local-bbp $(go env GOPATH)/bin/bbp
