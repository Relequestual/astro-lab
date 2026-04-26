package tui

import (
	"github.com/Relequestual/astro-lab/internal/models"
	"github.com/Relequestual/astro-lab/internal/storage"
	syncpkg "github.com/Relequestual/astro-lab/internal/sync"
)

// Auth messages

type authResolvedMsg struct {
	token    string
	provider models.AuthProvider
	err      error
}

type authValidatedMsg struct {
	login     string
	rateLimit *models.RateLimit
	err       error
}

// Data messages

type dataLoadedMsg struct {
	metadata    *models.Metadata
	stars       *storage.StarsData
	lists       *storage.ListsData
	memberships *storage.MembershipsData
	err         error
}

// Sync messages

type syncStartMsg struct{ full bool }

type syncProgressMsg struct {
	progress syncpkg.SyncProgress
}

type syncCompleteMsg struct {
	result *syncpkg.SyncResult
	err    error
}

// List mutation messages

type listCreatedMsg struct {
	list models.StarList
	err  error
}

type listUpdatedMsg struct {
	listID  string
	newName string
	err     error
}

type listDeletedMsg struct {
	listID string
	err    error
}

// List picker messages

type showListPickerMsg struct {
	repoID   string
	repoName string
}

type listPickerConfirmedMsg struct {
	repoID, repoName string
	selectedIDs       []string
	previousIDs       []string
}

// Mutation execution messages

type updateListsResultMsg struct {
	repoID, repoName string
	err               error
}

type undoResultMsg struct {
	description string
	err         error
}

// General messages

type statusMsg struct {
	text    string
	isError bool
}

type navigateMsg struct {
	screen           Screen
	listID, listName string
	repoID           string
}

type errMsg struct{ err error }

// backMsg requests navigation to the previous screen (handled by root model).
type backMsg struct{}

func (e errMsg) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return "unknown error"
}
