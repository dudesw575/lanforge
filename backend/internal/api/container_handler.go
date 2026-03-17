package api

import (
	"encoding/json"
	"net/http"

	"lanforge/internal/service"

	"github.com/go-chi/chi/v5"
)

type ContainerHandler struct {
	service *service.ContainerService
}

func NewContainerHandler(s *service.ContainerService) *ContainerHandler {
	return &ContainerHandler{service: s}
}

func (h *ContainerHandler) List(w http.ResponseWriter, r *http.Request) {
	containers, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(containers)
}

func (h *ContainerHandler) Start(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.service.Start(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}

func (h *ContainerHandler) Stop(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.service.Stop(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}

func (h *ContainerHandler) Remove(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.service.Remove(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}
