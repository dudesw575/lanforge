package api

import (
	"net/http"

	"lanforge/internal/auth"

	"lanforge/internal/websocket"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *ContainerHandler, authService *auth.Auth) *chi.Mux {

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {

		r.Use(authService.Middleware)

		r.Get("/containers", h.List)
		r.With(auth.RequireRole("admin", "operator")).Post("/containers/{id}/start", h.Start)
		r.With(auth.RequireRole("admin", "operator")).Post("/containers/{id}/stop", h.Stop)
		r.With(auth.RequireRole("admin")).Delete("/containers/{id}", h.Remove)

		r.Get("/containers/{id}/logs", func(w http.ResponseWriter, r *http.Request) {
			containerID := chi.URLParam(r, "id")
			websocket.StreamLogs(h.service.Docker, w, r, containerID)
		})

		r.Get("/containers/{id}/stats", func(w http.ResponseWriter, r *http.Request) {
			containerID := chi.URLParam(r, "id")
			websocket.StreamStats(h.service.Docker, w, r, containerID)
		})
	})
	return r
}
