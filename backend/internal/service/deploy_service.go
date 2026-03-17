package service

import (
	"context"
	"fmt"
	"lanforge/internal/templates"
	"lanforge/internal/websocket"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
)

type Deployment struct {
	TemplateID  string `json:"templateId"`
	ContainerID string `json:"containerId"`
	Name        string `json:"name"`
}

type DeployService struct {
	container   *ContainerService
	deployments map[string]*Deployment
}

func NewDeployService(containerSvc *ContainerService) *DeployService {
	return &DeployService{
		container:   containerSvc,
		deployments: make(map[string]*Deployment),
	}
}

func (d *DeployService) DeployTemplate(tmpl *templates.Template, name string) (*Deployment, error) {
	streamID := name

	broadcast := func(msg string) {
		websocket.DeployHub.Broadcast(streamID, msg)
	}

	broadcast("🚀 Deployment started")

	broadcast("📦 Preparing environment variables...")
	env := []string{}
	for k, v := range tmpl.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	broadcast("📊 Setting up ports...")
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for _, p := range tmpl.Ports {
		port := nat.Port(fmt.Sprintf("%d/tcp", p.Container))
		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{{HostPort: fmt.Sprintf("%d", p.Host)}}
	}

	broadcast("🗂 Creating and mounting volumes...")
	mounts := []mount.Mount{}
	for _, v := range tmpl.Volumes {
		volName := fmt.Sprintf("%s_data", name)

		_, err := d.container.CreateVolume(context.Background(), volName)
		if err != nil {
			msg := fmt.Sprintf("❌ Failed creating volume: %v", err)
			broadcast(msg)
			return nil, fmt.Errorf(msg)
		}

		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeVolume,
			Source: volName,
			Target: v.Container,
		})
	}

	broadcast("⬇ Pulling Docker image (this may take a minute)...")
	if err := d.container.PullImage(context.Background(), tmpl.Image); err != nil {
		msg := fmt.Sprintf("❌ Failed to pull image: %s", err)
		broadcast(msg)
		return nil, fmt.Errorf(msg)
	}
	broadcast("✅ Image pulled successfully")

	broadcast("📦 Creating container...")
	containerID, err := d.container.CreateAndStart(
		context.Background(),
		name,
		tmpl.Image,
		env,
		exposedPorts,
		portBindings,
		mounts,
		tmpl.Restart,
	)
	if err != nil {
		msg := fmt.Sprintf("❌ Container start failed: %s", err)
		broadcast(msg)
		return nil, fmt.Errorf(msg)
	}

	broadcast("🎉 Container started successfully")

	deployment := &Deployment{
		TemplateID:  tmpl.ID,
		ContainerID: containerID,
		Name:        name,
	}
	d.deployments[name] = deployment

	broadcast("✔ Deployment complete")

	return deployment, nil
}
