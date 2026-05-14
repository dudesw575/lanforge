package api

import (
	"net/http"

	"lanforge/internal/auth"
	"lanforge/internal/websocket"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	containerHandler *ContainerHandler,
	templateHandler *TemplateHandler,
	authService *auth.Auth,
) *chi.Mux {

	r := chi.NewRouter()

	// Deploy stream — public WebSocket endpoint (stream ID is a random token, not sensitive)
	r.Get("/deploy/stream", websocket.StreamDeployProgress)

	r.Group(func(r chi.Router) {
		r.Use(authService.Middleware)

		r.Get("/containers", containerHandler.List)
		r.With(auth.RequireRole("admin", "operator")).Post("/containers/{id}/start", containerHandler.Start)
		r.With(auth.RequireRole("admin", "operator")).Post("/containers/{id}/stop", containerHandler.Stop)
		r.With(auth.RequireRole("admin")).Delete("/containers/{id}", containerHandler.Remove)

		r.Get("/containers/{id}/logs", func(w http.ResponseWriter, r *http.Request) {
			containerID := chi.URLParam(r, "id")
			websocket.StreamLogs(containerHandler.service.Docker, w, r, containerID)
		})

		r.Get("/containers/{id}/stats", func(w http.ResponseWriter, r *http.Request) {
			containerID := chi.URLParam(r, "id")
			websocket.StreamStats(containerHandler.service.Docker, w, r, containerID)
		})

		r.Get("/templates", templateHandler.List)
		r.Get("/templates/{id}", templateHandler.Get)
		r.With(auth.RequireRole("admin", "operator")).Post("/templates", templateHandler.Create)
		r.With(auth.RequireRole("admin", "operator")).Post("/templates/{id}/deploy", templateHandler.Deploy)
		r.With(auth.RequireRole("admin")).Put("/templates/{id}", templateHandler.Update)
		r.With(auth.RequireRole("admin")).Delete("/templates/{id}", templateHandler.Delete)

		volumeHandler := NewVolumeHandler(containerHandler.service)
		r.Get("/volumes", volumeHandler.List)
		r.Post("/volumes/{name}", volumeHandler.Create)
		r.Delete("/volumes/{name}", volumeHandler.Delete)
	})

	return r
}
