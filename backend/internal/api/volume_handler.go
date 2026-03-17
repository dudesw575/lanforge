package api

import (
	"encoding/json"
	"net/http"

	"lanforge/internal/service"

	"github.com/go-chi/chi/v5"
)

type VolumeHandler struct {
	service *service.ContainerService
}

func NewVolumeHandler(s *service.ContainerService) *VolumeHandler {
	return &VolumeHandler{service: s}
}

func (h *VolumeHandler) List(w http.ResponseWriter, r *http.Request) {
	vols, err := h.service.ListVolumes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(vols)
}

func (h *VolumeHandler) Create(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	volName, err := h.service.CreateVolume(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"name": volName})
}

func (h *VolumeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if err := h.service.RemoveVolume(r.Context(), name, true); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
