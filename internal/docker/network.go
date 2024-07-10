package docker

import (
	"context"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type Network struct {
	ID         string
	Name       string
	client     *client.Client
	Containers []*Container
}

func NewNetwork(name string) *Network {
	return &Network{Name: name, client: dockerClient}
}

func (n *Network) Create(ctx context.Context) error {
	resp, err := n.client.NetworkCreate(ctx, n.Name, network.CreateOptions{
		Driver: "bridge",
	})
	if err != nil {
		return err
	}
	n.ID = resp.ID
	return nil
}

func (n *Network) AddService(c *Container) {
	n.Containers = append(n.Containers, c)
}

func (n *Network) Destroy(ctx context.Context) error {
	for _, c := range n.Containers {
		if err := c.Destroy(ctx); err != nil {
			return err
		}
	}
	n.Containers = nil
	return n.client.NetworkRemove(ctx, n.ID)
}
