package api

import (
	"encoding/json"
	"io"
	"lanforge/internal/service"
	"lanforge/internal/templates"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
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

func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tmpl, ok := h.store.Get(id)
	if !ok {
		http.Error(w, "template not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(tmpl)
}

// Create accepts a template JSON body or a YAML file upload
func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	var tmpl templates.Template

	switch {
	case contentType == "application/json":
		if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
			http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}

	case contentType == "application/x-yaml" || contentType == "text/yaml":
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			http.Error(w, "invalid YAML: "+err.Error(), http.StatusBadRequest)
			return
		}

	default:
		// multipart file upload (YAML file)
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "expected a YAML file upload (multipart field 'file') or JSON/YAML body", http.StatusBadRequest)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "failed to read uploaded file", http.StatusBadRequest)
			return
		}
		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			http.Error(w, "invalid YAML in uploaded file: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if tmpl.ID == "" || tmpl.Name == "" || tmpl.Image == "" {
		http.Error(w, "template must have 'id', 'name', and 'image' fields", http.StatusBadRequest)
		return
	}

	if err := h.store.Create(&tmpl); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tmpl)
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	contentType := r.Header.Get("Content-Type")

	var bodyTpl templates.Template

	switch {
	case contentType == "application/json":
		if err := json.NewDecoder(r.Body).Decode(&bodyTpl); err != nil {
			http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}

	default:
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		// Try JSON first, then YAML
		if err := json.Unmarshal(data, &bodyTpl); err != nil {
			if err := yaml.Unmarshal(data, &bodyTpl); err != nil {
				http.Error(w, "expected JSON or YAML body", http.StatusBadRequest)
				return
			}
		}
	}

	// Override the ID from the URL so the file path stays consistent
	bodyTpl.ID = id

	if bodyTpl.Name == "" || bodyTpl.Image == "" {
		http.Error(w, "template must have 'name' and 'image' fields", http.StatusBadRequest)
		return
	}

	if err := h.store.Update(&bodyTpl); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bodyTpl)
}

func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.store.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
