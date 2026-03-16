package main

import (
	"log"
	"net/http"

	"lanforge/internal/api"
	"lanforge/internal/auth"
	"lanforge/internal/docker"
	"lanforge/internal/service"
)

func main() {

	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	authService, err := auth.NewAuth(
		"http://localhost:8081/realms/LanParty",
	)

	if err != nil {
		log.Fatal(err)
	}

	containerService := service.NewContainerService(dockerClient)
	handler := api.NewContainerHandler(containerService)

	router := api.NewRouter(handler, authService)

	log.Println("server started on :8080")

	http.ListenAndServe(":8080", router)
}
