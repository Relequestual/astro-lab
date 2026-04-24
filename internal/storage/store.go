package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Relequestual/astro-lab/internal/models"
)

// Store manages local JSON file persistence
type Store struct {
	dir string
}

// NewStore creates a store at the given directory
func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// DefaultDir returns the default storage directory
func DefaultDir() string {
	if d := os.Getenv("ASTLAB_DATA_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		// Fall back to current directory if home is unavailable
		home = "."
	}
	return filepath.Join(home, ".astlab")
}

func (s *Store) path(name string) string {
	return filepath.Join(s.dir, name)
}

// LoadMetadata reads metadata.json
func (s *Store) LoadMetadata() (*models.Metadata, error) {
	data, err := os.ReadFile(s.path("metadata.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &models.Metadata{SchemaVersion: models.CurrentSchemaVersion}, nil
		}
		return nil, fmt.Errorf("reading metadata: %w", err)
	}
	var m models.Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing metadata: %w", err)
	}
	return &m, nil
}

// SaveMetadata writes metadata.json atomically
func (s *Store) SaveMetadata(m *models.Metadata) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling metadata: %w", err)
	}
	return AtomicWrite(s.path("metadata.json"), data)
}

// StarsData holds the stars file structure
type StarsData struct {
	ByRepoID map[string]models.Repository `json:"byRepoId"`
}

// LoadStars reads stars.json
func (s *Store) LoadStars() (*StarsData, error) {
	data, err := os.ReadFile(s.path("stars.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &StarsData{ByRepoID: make(map[string]models.Repository)}, nil
		}
		return nil, fmt.Errorf("reading stars: %w", err)
	}
	var sd StarsData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, fmt.Errorf("parsing stars: %w", err)
	}
	if sd.ByRepoID == nil {
		sd.ByRepoID = make(map[string]models.Repository)
	}
	return &sd, nil
}

// SaveStars writes stars.json atomically
func (s *Store) SaveStars(sd *StarsData) error {
	data, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling stars: %w", err)
	}
	return AtomicWrite(s.path("stars.json"), data)
}

// ListsData holds the lists file structure
type ListsData struct {
	ByListID map[string]models.StarList `json:"byListId"`
}

// LoadLists reads lists.json
func (s *Store) LoadLists() (*ListsData, error) {
	data, err := os.ReadFile(s.path("lists.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &ListsData{ByListID: make(map[string]models.StarList)}, nil
		}
		return nil, fmt.Errorf("reading lists: %w", err)
	}
	var ld ListsData
	if err := json.Unmarshal(data, &ld); err != nil {
		return nil, fmt.Errorf("parsing lists: %w", err)
	}
	if ld.ByListID == nil {
		ld.ByListID = make(map[string]models.StarList)
	}
	return &ld, nil
}

// SaveLists writes lists.json atomically
func (s *Store) SaveLists(ld *ListsData) error {
	data, err := json.MarshalIndent(ld, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling lists: %w", err)
	}
	return AtomicWrite(s.path("lists.json"), data)
}

// MembershipsData holds the memberships file structure
type MembershipsData struct {
	ListToRepos map[string][]string `json:"listToRepos"`
	RepoToLists map[string][]string `json:"repoToLists"`
}

// LoadMemberships reads memberships.json
func (s *Store) LoadMemberships() (*MembershipsData, error) {
	data, err := os.ReadFile(s.path("memberships.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &MembershipsData{
				ListToRepos: make(map[string][]string),
				RepoToLists: make(map[string][]string),
			}, nil
		}
		return nil, fmt.Errorf("reading memberships: %w", err)
	}
	var md MembershipsData
	if err := json.Unmarshal(data, &md); err != nil {
		return nil, fmt.Errorf("parsing memberships: %w", err)
	}
	if md.ListToRepos == nil {
		md.ListToRepos = make(map[string][]string)
	}
	if md.RepoToLists == nil {
		md.RepoToLists = make(map[string][]string)
	}
	return &md, nil
}

// SaveMemberships writes memberships.json atomically
func (s *Store) SaveMemberships(md *MembershipsData) error {
	data, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling memberships: %w", err)
	}
	return AtomicWrite(s.path("memberships.json"), data)
}

// Clear removes all stored data
func (s *Store) Clear() error {
	files := []string{"metadata.json", "stars.json", "lists.json", "memberships.json"}
	for _, f := range files {
		p := s.path(f)
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing %s: %w", f, err)
		}
	}
	return nil
}
