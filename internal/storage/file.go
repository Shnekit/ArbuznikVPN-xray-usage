package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Snapshot struct {
	Upload    uint64    `json:"upload"`
	Download  uint64    `json:"download"`
	UpdatedAt int64     `json:"updated_at"`
}

type DayUsage struct {
	Upload   uint64 `json:"upload"`
	Download uint64 `json:"download"`
}

type Store struct {
	path string
	mu   sync.Mutex

	Snapshots map[string]Snapshot
	Daily     map[string]map[string]DayUsage
}

func Open(path string) (*Store, error) {

	s := &Store{
		path: path,
		Snapshots: map[string]Snapshot{},
		Daily:     map[string]map[string]DayUsage{},
	}

	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return s, nil
}

func (s *Store) snapshotFile() string {
	return filepath.Join(s.path, "snapshots.json")
}

func (s *Store) dailyFile() string {
	return filepath.Join(s.path, "daily.json")
}

func (s *Store) load() error {

	data, err := os.ReadFile(s.snapshotFile())
	if err == nil {
		json.Unmarshal(data, &s.Snapshots)
	}

	data, err = os.ReadFile(s.dailyFile())
	if err == nil {
		json.Unmarshal(data, &s.Daily)
	}

	return nil
}

func (s *Store) Save() error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.path, 0755); err != nil {
		return err
	}

	if err := writeJSON(s.snapshotFile(), s.Snapshots); err != nil {
		return err
	}

	if err := writeJSON(s.dailyFile(), s.Daily); err != nil {
		return err
	}

	return nil
}

func writeJSON(path string, v any) error {

	tmp := path + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		f.Close()
		return err
	}

	f.Close()

	return os.Rename(tmp, path)
}

// Get snapshot
func (s *Store) GetSnapshot(user string) (Snapshot, bool) {
	v, ok := s.Snapshots[user]
	return v, ok
}

// Update snapshot
func (s *Store) SetSnapshot(user string, snap Snapshot) {
	s.Snapshots[user] = snap
}

// Add daily usage
func (s *Store) AddDaily(day string, user string, up, down uint64) {

	if s.Daily[day] == nil {
		s.Daily[day] = map[string]DayUsage{}
	}

	u := s.Daily[day][user]
	u.Upload += up
	u.Download += down
	s.Daily[day][user] = u
}

// Cleanup old days (keep 30)
func (s *Store) Cleanup() {

	cutoff := time.Now().Add(-30 * 24 * time.Hour)

	for day := range s.Daily {

		t, err := time.Parse("2006-01-02", day)
		if err != nil {
			continue
		}

		if t.Before(cutoff) {
			delete(s.Daily, day)
		}
	}
}