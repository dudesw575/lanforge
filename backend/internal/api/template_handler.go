package api

import (
	"encoding/json"
	"lanforge/internal/service"
	"lanforge/internal/templates"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type TemplateHandler struct {
	store  *templates.Store
	deploy *service.DeployService
}

func NewTemplateHandler(store *templates.Store, deploy *service.DeployService) *TemplateHandler {
	return &TemplateHandler{store: store, deploy: deploy}
}

func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(h.store.List())
}

func (h *TemplateHandler) Deploy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	name := r.URL.Query().Get("name")

	tmpl, ok := h.store.Get(id)
	if !ok {
		http.Error(w, "template not found", http.StatusNotFound)
		return
	}
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	deployment, err := h.deploy.DeployTemplate(tmpl, name)
	if err != nil {
		log.Printf("deploy error: template=%s name=%s err=%v", id, name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployment)
}

func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tmpl, ok := h.store.Get(id)
	if !ok {
		http.Error(w, "template not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(tmpl)
}
