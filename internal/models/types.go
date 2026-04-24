package models

import "time"

type Repository struct {
	ID            string    `json:"id"`
	NameWithOwner string    `json:"nameWithOwner"`
	StarredAt     time.Time `json:"starredAt"`
	Description   string    `json:"description,omitempty"`
	URL           string    `json:"url,omitempty"`
}

type StarList struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	IsPrivate   bool      `json:"isPrivate"`
	ItemCount   int       `json:"itemCount"`
	UpdatedAt   time.Time `json:"updatedAt"`
	LastAddedAt time.Time `json:"lastAddedAt,omitempty"`
}

type Metadata struct {
	SchemaVersion  int       `json:"schemaVersion"`
	AccountLogin   string    `json:"accountLogin"`
	LastSyncedAt   time.Time `json:"lastSyncedAt,omitempty"`
	LastFullSyncAt time.Time `json:"lastFullSyncAt,omitempty"`
}

type Membership struct {
	ListID string `json:"listId"`
	RepoID string `json:"repoId"`
}

type MoveOperation struct {
	RepoID          string   `json:"repoId"`
	RepoName        string   `json:"repoName"`
	AddToLists      []string `json:"addToLists,omitempty"`
	RemoveFromLists []string `json:"removeFromLists,omitempty"`
}

type MoveDiff struct {
	Before  []string `json:"before"`
	After   []string `json:"after"`
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
}

// Auth types
type AuthProvider string

const (
	AuthProviderToken   AuthProvider = "token"
	AuthProviderGH      AuthProvider = "gh"
	AuthProviderEnv     AuthProvider = "env"
	AuthProviderKeyring AuthProvider = "keyring"
)

type AuthStore string

const (
	AuthStoreKeyring AuthStore = "keyring"
	AuthStoreNone    AuthStore = "none"
)

type AuthStatus struct {
	Provider      AuthProvider `json:"provider"`
	Login         string       `json:"login"`
	Authenticated bool         `json:"authenticated"`
}

const CurrentSchemaVersion = 1
