package docker

import (
	"context"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type Network struct {
	ID     string
	Name   string
	client *client.Client
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

func (n *Network) Destroy(ctx context.Context) error {
	return n.client.NetworkRemove(ctx, n.ID)
}
