package docker

import (
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

var dockerClient *client.Client
var err error

func init() {
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
}
