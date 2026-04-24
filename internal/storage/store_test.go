package storage

import (
	"os"
	"testing"
	"time"

	"github.com/Relequestual/astro-lab/internal/models"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "astlab-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestMetadata_RoundTrip(t *testing.T) {
	dir := tempDir(t)
	store := NewStore(dir)

	meta := &models.Metadata{
		SchemaVersion: 1,
		AccountLogin:  "testuser",
		LastSyncedAt:  time.Now().UTC().Truncate(time.Second),
	}

	if err := store.SaveMetadata(meta); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := store.LoadMetadata()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.SchemaVersion != meta.SchemaVersion {
		t.Errorf("schemaVersion: got %d want %d", loaded.SchemaVersion, meta.SchemaVersion)
	}
	if loaded.AccountLogin != meta.AccountLogin {
		t.Errorf("accountLogin: got %q want %q", loaded.AccountLogin, meta.AccountLogin)
	}
}

func TestStars_RoundTrip(t *testing.T) {
	dir := tempDir(t)
	store := NewStore(dir)

	stars := &StarsData{
		ByRepoID: map[string]models.Repository{
			"R_123": {
				ID:            "R_123",
				NameWithOwner: "test/repo",
				StarredAt:     time.Now().UTC().Truncate(time.Second),
			},
		},
	}

	if err := store.SaveStars(stars); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := store.LoadStars()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(loaded.ByRepoID) != 1 {
		t.Fatalf("expected 1 star, got %d", len(loaded.ByRepoID))
	}

	repo := loaded.ByRepoID["R_123"]
	if repo.NameWithOwner != "test/repo" {
		t.Errorf("nameWithOwner: got %q want %q", repo.NameWithOwner, "test/repo")
	}
}

func TestLists_RoundTrip(t *testing.T) {
	dir := tempDir(t)
	store := NewStore(dir)

	lists := &ListsData{
		ByListID: map[string]models.StarList{
			"UL_1": {
				ID:        "UL_1",
				Name:      "Test List",
				Slug:      "test-list",
				ItemCount: 5,
			},
		},
	}

	if err := store.SaveLists(lists); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := store.LoadLists()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(loaded.ByListID) != 1 {
		t.Fatalf("expected 1 list, got %d", len(loaded.ByListID))
	}

	list := loaded.ByListID["UL_1"]
	if list.Name != "Test List" {
		t.Errorf("name: got %q want %q", list.Name, "Test List")
	}
}

func TestMemberships_RoundTrip(t *testing.T) {
	dir := tempDir(t)
	store := NewStore(dir)

	memberships := &MembershipsData{
		ListToRepos: map[string][]string{
			"UL_1": {"R_1", "R_2"},
		},
		RepoToLists: map[string][]string{
			"R_1": {"UL_1"},
			"R_2": {"UL_1"},
		},
	}

	if err := store.SaveMemberships(memberships); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := store.LoadMemberships()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(loaded.ListToRepos["UL_1"]) != 2 {
		t.Errorf("expected 2 repos in list, got %d", len(loaded.ListToRepos["UL_1"]))
	}
}

func TestLoadMetadata_MissingFile(t *testing.T) {
	dir := tempDir(t)
	store := NewStore(dir)

	meta, err := store.LoadMetadata()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.SchemaVersion != models.CurrentSchemaVersion {
		t.Errorf("expected default schema version %d, got %d", models.CurrentSchemaVersion, meta.SchemaVersion)
	}
}

func TestClear(t *testing.T) {
	dir := tempDir(t)
	store := NewStore(dir)

	// Create some files
	store.SaveMetadata(&models.Metadata{SchemaVersion: 1})
	store.SaveStars(&StarsData{ByRepoID: map[string]models.Repository{}})

	if err := store.Clear(); err != nil {
		t.Fatalf("clear: %v", err)
	}

	// Verify files are gone (load should return defaults)
	meta, _ := store.LoadMetadata()
	if meta.AccountLogin != "" {
		t.Error("expected empty account login after clear")
	}
}
