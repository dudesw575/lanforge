package templates

import (
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
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".yaml" {
			continue
		}
		data, _ := os.ReadFile(filepath.Join(dir, f.Name()))
		var t Template
		if err := yaml.Unmarshal(data, &t); err != nil {
			continue
		}
		s.templates[t.ID] = &t
	}
	return nil
}

func (s *Store) List() []*Template {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := []*Template{}
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
