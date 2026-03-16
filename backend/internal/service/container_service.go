package service

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ContainerService struct {
	Docker *client.Client
}

func NewContainerService(docker *client.Client) *ContainerService {
	return &ContainerService{Docker: docker}
}

// List all containers
func (s *ContainerService) List(ctx context.Context) ([]container.Summary, error) {
	return s.Docker.ContainerList(ctx, container.ListOptions{
		All: true,
	})
}

// Start a container
func (s *ContainerService) Start(ctx context.Context, id string) error {
	return s.Docker.ContainerStart(ctx, id, container.StartOptions{})
}

// Stop a container
func (s *ContainerService) Stop(ctx context.Context, id string) error {
	return s.Docker.ContainerStop(ctx, id, container.StopOptions{})
}

// Remove a container
func (s *ContainerService) Remove(ctx context.Context, id string) error {
	return s.Docker.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}
