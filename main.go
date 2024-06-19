package main

import (
	_ "embed"
	"github/zhex/bbp/cmd"
	"log"
)

//go:embed VERSION
var version string

func main() {
	c := cmd.CreateRootCmd(version)
	if err := c.Execute(); err != nil {
		log.Fatal(err)
	}
}
