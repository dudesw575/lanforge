package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type Store struct {
	mu        sync.RWMutex
	templates map[string]*Template
	dir       string
}

func NewStore(dir string) (*Store, error) {
	s := &Store{
		templates: make(map[string]*Template),
		dir:       dir,
	}
	if err := s.loadDirectory(dir); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) loadDirectory(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".yaml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			continue
		}
		var t Template
		if err := yaml.Unmarshal(data, &t); err != nil {
			continue
		}
		t.Normalize()
		s.templates[t.ID] = &t
	}
	return nil
}

func (s *Store) List() []*Template {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Template, 0, len(s.templates))
	for _, t := range s.templates {
		list = append(list, t)
	}
	return list
}

func (s *Store) Get(id string) (*Template, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.templates[id]
	return t, ok
}

func (s *Store) Create(t *Template) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	t.Normalize()

	if _, exists := s.templates[t.ID]; exists {
		return fmt.Errorf("template %q already exists", t.ID)
	}

	data, err := yaml.Marshal(t)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	path := filepath.Join(s.dir, t.ID+".yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	s.templates[t.ID] = t
	return nil
}

func (s *Store) Update(t *Template) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	t.Normalize()

	if _, exists := s.templates[t.ID]; !exists {
		return fmt.Errorf("template %q not found", t.ID)
	}

	data, err := yaml.Marshal(t)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	path := filepath.Join(s.dir, t.ID+".yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	s.templates[t.ID] = t
	return nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.templates[id]; !exists {
		return fmt.Errorf("template %q not found", id)
	}

	path := filepath.Join(s.dir, id+".yaml")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove template file: %w", err)
	}

	delete(s.templates, id)
	return nil
}
