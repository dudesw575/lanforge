package service

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerimage "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type ContainerService struct {
	Docker *client.Client
}

func NewContainerService(docker *client.Client) *ContainerService {
	return &ContainerService{Docker: docker}
}

func (s *ContainerService) List(ctx context.Context) ([]container.Summary, error) {
	return s.Docker.ContainerList(ctx, container.ListOptions{All: true})
}

func (s *ContainerService) Start(ctx context.Context, id string) error {
	return s.Docker.ContainerStart(ctx, id, container.StartOptions{})
}

func (s *ContainerService) Stop(ctx context.Context, id string) error {
	return s.Docker.ContainerStop(ctx, id, container.StopOptions{})
}

func (s *ContainerService) Remove(ctx context.Context, id string) error {
	return s.Docker.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}

func (s *ContainerService) PullImage(ctx context.Context, image string) error {
	reader, err := s.Docker.ImagePull(ctx, image, dockerimage.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %q: %w", image, err)
	}
	defer reader.Close()

	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return fmt.Errorf("reading pull response for %q: %w", image, err)
	}

	return nil
}

func (s *ContainerService) ListVolumes(ctx context.Context) ([]string, error) {
	resp, err := s.Docker.VolumeList(ctx, volume.ListOptions{Filters: filters.Args{}})
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, vol := range resp.Volumes {
		names = append(names, vol.Name)
	}
	return names, nil
}

func (s *ContainerService) CreateVolume(ctx context.Context, name string) (string, error) {
	vol, err := s.Docker.VolumeCreate(ctx, volume.CreateOptions{
		Name: name,
	})
	if err != nil {
		return "", err
	}
	return vol.Name, nil
}

func (s *ContainerService) RemoveVolume(ctx context.Context, name string, force bool) error {
	return s.Docker.VolumeRemove(ctx, name, force)
}

func (s *ContainerService) CreateAndStart(
	ctx context.Context,
	name string,
	image string,
	env []string,
	exposedPorts nat.PortSet,
	portBindings nat.PortMap,
	volumes []mount.Mount,
	restart string,
) (string, error) {

	if err := s.PullImage(ctx, image); err != nil {
		return "", err
	}

	config := &container.Config{
		Image:        image,
		Env:          env,
		ExposedPorts: exposedPorts,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Mounts:       volumes,
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyMode(restart),
		},
	}

	resp, err := s.Docker.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("container create failed: %w", err)
	}

	if err := s.Docker.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("container start failed: %w", err)
	}

	return resp.ID, nil
}
