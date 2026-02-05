package db

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type State struct {
	Users   []map[string]any `json:"users"`
	Servers []map[string]any `json:"servers"`
	Tasks   []map[string]any `json:"tasks"`
	Seq     map[string]int64 `json:"seq"`
}

type FileDB struct {
	mu    sync.Mutex
	path  string
	state State
}

func OpenFileDB(path string) (*FileDB, error) {
	d := &FileDB{path: path}
	if err := d.load(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *FileDB) load() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(d.path), 0o755); err != nil {
		return err
	}
	b, err := os.ReadFile(d.path)
	if err != nil {
		if os.IsNotExist(err) {
			d.state = State{Users: []map[string]any{}, Servers: []map[string]any{}, Tasks: []map[string]any{}, Seq: map[string]int64{}}
			return d.flushLocked()
		}
		return err
	}
	if len(b) == 0 {
		d.state = State{Users: []map[string]any{}, Servers: []map[string]any{}, Tasks: []map[string]any{}, Seq: map[string]int64{}}
		return nil
	}
	if err := json.Unmarshal(b, &d.state); err != nil {
		return err
	}
	if d.state.Seq == nil {
		d.state.Seq = map[string]int64{}
	}
	return nil
}

func (d *FileDB) Snapshot() State {
	d.mu.Lock()
	defer d.mu.Unlock()
	return cloneState(d.state)
}

func (d *FileDB) Update(fn func(*State) error) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := fn(&d.state); err != nil {
		return err
	}
	return d.flushLocked()
}

func (d *FileDB) flushLocked() error {
	b, err := json.MarshalIndent(d.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(d.path, b, 0o644)
}

func cloneState(s State) State {
	clone := State{Seq: map[string]int64{}}
	for k, v := range s.Seq {
		clone.Seq[k] = v
	}
	clone.Users = append([]map[string]any(nil), s.Users...)
	clone.Servers = append([]map[string]any(nil), s.Servers...)
	clone.Tasks = append([]map[string]any(nil), s.Tasks...)
	return clone
}
