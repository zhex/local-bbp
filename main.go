package main

import (
	_ "embed"
	log "github.com/sirupsen/logrus"
	"github/zhex/local-bbp/cmd"
)

//go:embed VERSION
var version string

func main() {
	log.SetLevel(log.DebugLevel)
	c := cmd.CreateRootCmd(version)
	if err := c.Execute(); err != nil {
		log.Fatal(err)
	}
}
