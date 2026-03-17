package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lanforge/internal/api"
	"lanforge/internal/auth"
	"lanforge/internal/docker"
	"lanforge/internal/service"
	"lanforge/internal/templates"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	dockerCli, err := docker.NewClient()
	if err != nil {
		log.Fatalf("Failed to connect to Docker: %v", err)
	}

	templateStore, err := templates.NewStore("./templates")
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	containerSvc := service.NewContainerService(dockerCli)
	deploySvc := service.NewDeployService(containerSvc)
	authSvc, err := auth.NewAuth(os.Getenv("OIDC_ISSUER"))
	if err != nil {
		log.Fatalf("Failed to init auth: %v", err)
	}

	containerH := api.NewContainerHandler(containerSvc)
	templateH := api.NewTemplateHandler(templateStore, deploySvc)

	router := api.NewRouter(containerH, templateH, authSvc)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("LanForge Backend starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	<-done
	log.Println("Server stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
