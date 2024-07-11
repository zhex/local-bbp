package main

import (
	_ "embed"
	log "github.com/sirupsen/logrus"
	"github.com/zhex/local-bbp/cmd"
)

//go:embed VERSION
var version string

func main() {
	c := cmd.CreateRootCmd(version)
	if err := c.Execute(); err != nil {
		log.Fatal(err)
	}
}
