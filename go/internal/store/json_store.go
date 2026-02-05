package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"aigcpanel/go/internal/domain"
)

type JSONStore struct {
	mu   sync.RWMutex
	path string
	db   domain.Database
}

func NewJSONStore(path string) (*JSONStore, error) {
	s := &JSONStore{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *JSONStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	buf, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.db = domain.Database{}
			return s.flushLocked()
		}
		return err
	}

	if len(buf) == 0 {
		s.db = domain.Database{}
		return nil
	}

	if err := json.Unmarshal(buf, &s.db); err != nil {
		return err
	}
	return nil
}

func (s *JSONStore) Snapshot() domain.Database {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db
}

func (s *JSONStore) Update(fn func(*domain.Database) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := fn(&s.db); err != nil {
		return err
	}
	return s.flushLocked()
}

func (s *JSONStore) flushLocked() error {
	buf, err := json.MarshalIndent(s.db, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, buf, 0o644)
}
