package docker

import "github.com/zhex/local-bbp/internal/models"

type Input struct {
	Name         string
	NetworkAlias string
	Image        *models.Image
	WorkDir      string
	HostDir      string
	Envs         map[string]string
	Entrypoint   []string
}
