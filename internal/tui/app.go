package tui

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Relequestual/astro-lab/internal/auth"
	"github.com/Relequestual/astro-lab/internal/github"
	"github.com/Relequestual/astro-lab/internal/models"
	"github.com/Relequestual/astro-lab/internal/storage"
	syncpkg "github.com/Relequestual/astro-lab/internal/sync"
)

const defaultUndoStackSize = 50

// Model is the root Bubble Tea model that wires all screens together.
type Model struct {
	screen     Screen
	prevScreen Screen

	token string
	login string

	rateLimit *models.RateLimit
	authProv  *auth.Provider
	store     *storage.Store
	client    *github.Client
	syncEng   *syncpkg.Engine
	program   *tea.Program

	metadata    *models.Metadata
	stars       *storage.StarsData
	lists       *storage.ListsData
	memberships *storage.MembershipsData

	width, height int

	statusText    string
	statusIsError bool
	showHelp      bool
	quitting      bool

	authScreen  authModel
	dashboard   dashboardModel
	listsPanel  listsModel
	starsPanel  starsModel
	reposInList reposInListModel
	detailPanel detailModel

	syncOverlay syncOverlayModel
	preview     previewModel
	listPicker  listPickerModel

	undoStack       *UndoStack
	pendingMutation *pendingMutation
}

type pendingMutation struct {
	repoID, repoName       string
	newListIDs, prevListIDs []string
}

// NewModel creates the root model with initial sub-models.
func NewModel(store *storage.Store, authProv *auth.Provider) Model {
	return Model{
		screen:      ScreenAuth,
		authProv:    authProv,
		store:       store,
		authScreen:  newAuthModel(authProv),
		dashboard:   newDashboardModel(),
		listsPanel:  newListsModel(),
		starsPanel:  newStarsModel(),
		reposInList: newReposInListModel("", ""),
		detailPanel: newDetailModel(models.Repository{}, nil),
		syncOverlay: newSyncOverlayModel(),
		preview:     newPreviewModel(),
		listPicker:  newListPickerModel(),
		undoStack:   NewUndoStack(defaultUndoStackSize),
	}
}

// SetProgram sets the tea.Program reference for sending messages from commands.
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(resolveAuthCmd(m.authProv), m.authScreen.Init())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.listsPanel.width, m.listsPanel.height = msg.Width, msg.Height
		m.starsPanel.width, m.starsPanel.height = msg.Width, msg.Height
		m.reposInList.stars.width, m.reposInList.stars.height = msg.Width, msg.Height
		m.detailPanel.width, m.detailPanel.height = msg.Width, msg.Height
		m.listPicker.width, m.listPicker.height = msg.Width, msg.Height
		return m, nil

	case authResolvedMsg:
		if msg.err != nil {
			m.screen = ScreenAuth
			return m, nil
		}
		m.token = msg.token
		client := github.NewClient(msg.token)
		return m, validateTokenCmdFromToken(client)

	case authValidatedMsg:
		if msg.err != nil {
			m.screen = ScreenAuth
			m.authScreen.err = fmt.Sprintf("Authentication failed: %s", msg.err)
			m.authScreen.validating = false
			return m, nil
		}
		m.login = msg.login
		m.rateLimit = msg.rateLimit
		if m.token == "" {
			// Token was entered manually via the auth screen
			m.token = m.authScreen.textInput.Value()
		}
		m.client = github.NewClient(m.token)
		m.syncEng = syncpkg.NewEngine(m.client, m.store)
		return m, loadDataCmd(m.store)

	case dataLoadedMsg:
		if msg.err != nil {
			m.statusText = fmt.Sprintf("Error loading data: %s", msg.err)
			m.statusIsError = true
			m.screen = ScreenDashboard
			return m, nil
		}
		m.metadata = msg.metadata
		m.stars = msg.stars
		m.lists = msg.lists
		m.memberships = msg.memberships
		m.populateScreens()
		needsSync := m.metadata.LastSyncedAt.IsZero()
		m.dashboard.needsSync = needsSync
		if !m.metadata.LastSyncedAt.IsZero() {
			m.dashboard.lastSync = humanDuration(time.Since(m.metadata.LastSyncedAt))
		}
		m.screen = ScreenDashboard
		return m, nil

	case navigateMsg:
		m.statusText = ""
		m.prevScreen = m.screen
		m.screen = msg.screen
		switch msg.screen {
		case ScreenReposInList:
			m.reposInList = newReposInListModel(msg.listID, msg.listName)
			m.reposInList.stars.width, m.reposInList.stars.height = m.width, m.height
			if m.memberships != nil {
				repoIDs := m.memberships.ListToRepos[msg.listID]
				repos := m.reposForIDs(repoIDs)
				m.reposInList.setRepos(repos)
				m.reposInList.stars.repoToLists = m.memberships.RepoToLists
			}
		case ScreenRepoDetail:
			if m.stars != nil {
				if repo, ok := m.stars.ByRepoID[msg.repoID]; ok {
					listNames := m.listNamesForRepo(msg.repoID)
					m.detailPanel = newDetailModel(repo, listNames)
					m.detailPanel.width, m.detailPanel.height = m.width, m.height
				}
			}
		}
		return m, nil

	case syncStartMsg:
		if m.token == "" {
			m.statusText = "Cannot sync: not authenticated"
			m.statusIsError = true
			return m, nil
		}
		ctx, cancel := context.WithCancel(context.Background())
		m.syncOverlay.start(cancel)
		cmds = append(cmds, m.syncOverlay.Init())
		cmds = append(cmds, makeSyncCmd(m.token, m.store, msg.full, ctx, m.program))
		return m, tea.Batch(cmds...)

	case syncProgressMsg:
		var cmd tea.Cmd
		m.syncOverlay, cmd = m.syncOverlay.Update(msg)
		return m, cmd

	case syncCompleteMsg:
		m.syncOverlay.stop()
		if msg.err != nil {
			// Distinguish cancellation from real errors
			if errors.Is(msg.err, context.Canceled) || errors.Is(msg.err, context.DeadlineExceeded) {
				m.statusText = "Sync cancelled"
				m.statusIsError = false
				return m, nil
			}
			m.statusText = fmt.Sprintf("Sync failed: %s", msg.err)
			m.statusIsError = true
			return m, nil
		}
		m.statusText = msg.result.String()
		m.statusIsError = false
		return m, loadDataCmd(m.store)

	case listCreatedMsg:
		if msg.err != nil {
			m.statusText = fmt.Sprintf("List create error: %s", msg.err)
			m.statusIsError = true
			return m, nil
		}
		if m.client == nil {
			m.statusText = "Cannot create list: not authenticated"
			m.statusIsError = true
			return m, nil
		}
		name := msg.list.Name
		client := m.client
		return m, func() tea.Msg { return createListCmd(client, name) }

	case listUpdatedMsg:
		if msg.err != nil {
			m.statusText = fmt.Sprintf("List rename error: %s", msg.err)
			m.statusIsError = true
			return m, nil
		}
		if m.client == nil {
			m.statusText = "Cannot rename list: not authenticated"
			m.statusIsError = true
			return m, nil
		}
		listID, newName := msg.listID, msg.newName
		client := m.client
		return m, func() tea.Msg { return renameListCmd(client, listID, newName) }

	case listDeletedMsg:
		if msg.err != nil {
			m.statusText = fmt.Sprintf("List delete error: %s", msg.err)
			m.statusIsError = true
			return m, nil
		}
		if m.client == nil {
			m.statusText = "Cannot delete list: not authenticated"
			m.statusIsError = true
			return m, nil
		}
		listID := msg.listID
		client := m.client
		return m, func() tea.Msg { return deleteListCmd(client, listID) }

	case showListPickerMsg:
		var allLists []models.StarList
		if m.lists != nil {
			allLists = listSlice(m.lists.ByListID)
			sort.Slice(allLists, func(i, j int) bool {
				return allLists[i].Name < allLists[j].Name
			})
		}
		var currentIDs []string
		if m.memberships != nil {
			currentIDs = m.memberships.RepoToLists[msg.repoID]
		}
		m.listPicker.show(msg.repoID, msg.repoName, allLists, currentIDs)
		return m, nil

	case listPickerConfirmedMsg:
		diff := buildDiff(msg.previousIDs, msg.selectedIDs)
		if len(diff.Added) == 0 && len(diff.Removed) == 0 {
			m.statusText = "No changes"
			m.statusIsError = false
			return m, nil
		}
		listNames := m.listNameMap()
		m.preview.show(msg.repoName, diff, listNames)
		m.pendingMutation = &pendingMutation{
			repoID:      msg.repoID,
			repoName:    msg.repoName,
			newListIDs:  msg.selectedIDs,
			prevListIDs: msg.previousIDs,
		}
		return m, nil

	case updateListsResultMsg:
		if msg.err != nil {
			m.statusText = fmt.Sprintf("Update failed for %s: %s", msg.repoName, msg.err)
			m.statusIsError = true
		} else {
			m.statusText = fmt.Sprintf("Updated lists for %s", msg.repoName)
			m.statusIsError = false
		}
		return m, loadDataCmd(m.store)

	case undoResultMsg:
		if msg.err != nil {
			m.statusText = fmt.Sprintf("Undo failed (%s): %s", msg.description, msg.err)
			m.statusIsError = true
		} else {
			m.statusText = fmt.Sprintf("Undone: %s", msg.description)
			m.statusIsError = false
		}
		return m, loadDataCmd(m.store)

	case statusMsg:
		m.statusText = msg.text
		m.statusIsError = msg.isError
		return m, nil

	case errMsg:
		m.statusText = msg.Error()
		m.statusIsError = true
		return m, nil

	case backMsg:
		return m, m.navigateBack()

	case tea.KeyMsg:
		// Overlay priority: sync overlay absorbs esc
		if m.syncOverlay.active {
			var cmd tea.Cmd
			m.syncOverlay, cmd = m.syncOverlay.Update(msg)
			return m, cmd
		}

		// Preview overlay
		if m.preview.active {
			var cmd tea.Cmd
			m.preview, cmd = m.preview.Update(msg)
			if m.preview.resolved {
				confirmed := m.preview.confirmed
				m.preview.hide()
				if confirmed && m.pendingMutation != nil {
					pm := m.pendingMutation
					m.pendingMutation = nil
					m.undoStack.Push(UndoEntry{
						Description: fmt.Sprintf("update lists for %s", pm.repoName),
						RepoID:      pm.repoID,
						RepoName:    pm.repoName,
						PreviousIDs: pm.prevListIDs,
					})
					return m, executeMutationCmd(m.client, m.store, m.memberships, pm.repoID, pm.repoName, pm.newListIDs)
				}
				m.pendingMutation = nil
			}
			return m, cmd
		}

		// List picker overlay
		if m.listPicker.active {
			var cmd tea.Cmd
			m.listPicker, cmd = m.listPicker.Update(msg)
			return m, cmd
		}

		// Help overlay
		if m.showHelp {
			if msg.String() == "?" || msg.String() == "esc" {
				m.showHelp = false
			}
			return m, nil
		}

		// Global keys
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "?":
			m.showHelp = true
			return m, nil
		case "q":
			if m.screen == ScreenAuth {
				m.quitting = true
				return m, tea.Quit
			}
			if m.screen == ScreenDashboard {
				m.quitting = true
				return m, tea.Quit
			}
			// On sub-screens, treat q as quit too
			m.quitting = true
			return m, tea.Quit
		case "esc":
			return m, m.navigateBack()
		case "s":
			if m.screen != ScreenAuth {
				return m, func() tea.Msg { return syncStartMsg{full: false} }
			}
		case "u":
			if m.screen != ScreenAuth {
				return m, m.handleUndo()
			}
		}

		// Delegate to active screen
		return m.updateActiveScreen(msg)
	}

	// Non-key messages: delegate spinner ticks etc. to overlays, then active screen
	if m.syncOverlay.active {
		var cmd tea.Cmd
		m.syncOverlay, cmd = m.syncOverlay.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	updated, cmd := m.updateActiveScreen(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	return updated, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.showHelp {
		return helpView(m.width, m.height)
	}

	var body string
	switch m.screen {
	case ScreenAuth:
		return m.authScreen.View()
	case ScreenDashboard:
		body = m.dashboard.View()
	case ScreenLists:
		body = m.listsPanel.View()
	case ScreenAllStars:
		body = m.starsPanel.View()
	case ScreenReposInList:
		body = m.reposInList.View()
	case ScreenRepoDetail:
		body = m.detailPanel.View()
	}

	// Overlays
	if m.syncOverlay.active {
		body = m.syncOverlay.View()
	}
	if m.preview.active {
		body = m.preview.View()
	}
	if m.listPicker.active {
		body = m.listPicker.View()
	}

	var lastSync time.Time
	if m.metadata != nil {
		lastSync = m.metadata.LastSyncedAt
	}
	header := headerView(m.width, m.login, rateLimitString(m.rateLimit), lastSync)
	footer := footerView(m.width, globalBindings(), m.statusText, m.statusIsError)

	return header + "\n" + body + "\n" + footer
}

// updateActiveScreen delegates a message to the currently active screen.
func (m *Model) updateActiveScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.screen {
	case ScreenAuth:
		m.authScreen, cmd = m.authScreen.Update(msg)
	case ScreenDashboard:
		m.dashboard, cmd = m.dashboard.Update(msg)
	case ScreenLists:
		m.listsPanel, cmd = m.listsPanel.Update(msg)
	case ScreenAllStars:
		m.starsPanel, cmd = m.starsPanel.Update(msg)
	case ScreenReposInList:
		m.reposInList, cmd = m.reposInList.Update(msg)
	case ScreenRepoDetail:
		m.detailPanel, cmd = m.detailPanel.Update(msg)
	}
	return m, cmd
}

// navigateBack returns a command to go to the previous screen.
func (m *Model) navigateBack() tea.Cmd {
	switch m.screen {
	case ScreenDashboard:
		return nil
	case ScreenLists, ScreenAllStars:
		return func() tea.Msg { return navigateMsg{screen: ScreenDashboard} }
	case ScreenReposInList:
		return func() tea.Msg { return navigateMsg{screen: ScreenLists} }
	case ScreenRepoDetail:
		prev := m.prevScreen
		if prev == ScreenAuth || prev == ScreenRepoDetail {
			prev = ScreenDashboard
		}
		return func() tea.Msg { return navigateMsg{screen: prev} }
	}
	return nil
}

// handleUndo pops from the undo stack and reverts the mutation.
func (m *Model) handleUndo() tea.Cmd {
	entry, ok := m.undoStack.Pop()
	if !ok {
		return func() tea.Msg { return statusMsg{text: "Nothing to undo"} }
	}
	if m.client == nil {
		return func() tea.Msg {
			return statusMsg{text: "Cannot undo: not authenticated", isError: true}
		}
	}
	desc := entry.Description
	repoID := entry.RepoID
	prevIDs := entry.PreviousIDs
	client := m.client
	memberships := m.memberships
	store := m.store
	return func() tea.Msg {
		ctx := context.Background()
		err := client.UpdateUserListsForItem(ctx, repoID, prevIDs)
		if err != nil {
			return undoResultMsg{description: desc, err: err}
		}
		if memberships != nil {
			updateLocalMemberships(memberships, repoID, prevIDs)
			if store != nil {
				if err := store.SaveMemberships(memberships); err != nil {
					return undoResultMsg{description: desc, err: fmt.Errorf("reverted remotely but failed to save locally: %w", err)}
				}
			}
		}
		return undoResultMsg{description: desc}
	}
}

// populateScreens fills sub-models with loaded data.
func (m *Model) populateScreens() {
	if m.stars != nil {
		repos := repoSlice(m.stars.ByRepoID)
		m.starsPanel.setRepos(repos)
		m.dashboard.totalStars = len(m.stars.ByRepoID)
	}
	if m.lists != nil {
		lists := listSlice(m.lists.ByListID)
		m.listsPanel.setLists(lists)
		m.dashboard.totalLists = len(m.lists.ByListID)
	}
	if m.memberships != nil {
		m.starsPanel.repoToLists = m.memberships.RepoToLists
	}
}

// reposForIDs returns Repository slices for a list of repo IDs.
func (m *Model) reposForIDs(ids []string) []models.Repository {
	if m.stars == nil {
		return nil
	}
	repos := make([]models.Repository, 0, len(ids))
	for _, id := range ids {
		if r, ok := m.stars.ByRepoID[id]; ok {
			repos = append(repos, r)
		}
	}
	return repos
}

// listNamesForRepo returns the list names a repo belongs to.
func (m *Model) listNamesForRepo(repoID string) []string {
	if m.memberships == nil || m.lists == nil {
		return nil
	}
	listIDs := m.memberships.RepoToLists[repoID]
	names := make([]string, 0, len(listIDs))
	for _, lid := range listIDs {
		if l, ok := m.lists.ByListID[lid]; ok {
			names = append(names, l.Name)
		}
	}
	return names
}

// listNameMap returns a map from list ID to list name.
func (m *Model) listNameMap() map[string]string {
	names := make(map[string]string)
	if m.lists != nil {
		for id, l := range m.lists.ByListID {
			names[id] = l.Name
		}
	}
	return names
}

// --- Commands ---

// resolveAuthCmd attempts to resolve a token from the auth provider.
func resolveAuthCmd(prov *auth.Provider) tea.Cmd {
	return func() tea.Msg {
		if prov == nil {
			return authResolvedMsg{err: fmt.Errorf("no auth provider")}
		}
		token, provider, err := prov.Resolve()
		if err != nil {
			return authResolvedMsg{err: err}
		}
		return authResolvedMsg{token: token, provider: provider}
	}
}

// validateTokenCmdFromToken validates a pre-resolved token.
func validateTokenCmdFromToken(client *github.Client) tea.Cmd {
	return func() tea.Msg {
		login, rl, err := client.ViewerLoginWithRateLimit(context.Background())
		if err != nil {
			return authValidatedMsg{err: err}
		}
		return authValidatedMsg{login: login, rateLimit: rl}
	}
}

// loadDataCmd loads all data from disk.
func loadDataCmd(store *storage.Store) tea.Cmd {
	return func() tea.Msg {
		meta, err := store.LoadMetadata()
		if err != nil {
			return dataLoadedMsg{err: err}
		}
		stars, err := store.LoadStars()
		if err != nil {
			return dataLoadedMsg{err: err}
		}
		lists, err := store.LoadLists()
		if err != nil {
			return dataLoadedMsg{err: err}
		}
		memberships, err := store.LoadMemberships()
		if err != nil {
			return dataLoadedMsg{err: err}
		}
		return dataLoadedMsg{
			metadata:    meta,
			stars:       stars,
			lists:       lists,
			memberships: memberships,
		}
	}
}

// executeMutationCmd calls UpdateUserListsForItem and updates local state.
func executeMutationCmd(client *github.Client, store *storage.Store, memberships *storage.MembershipsData, repoID, repoName string, newListIDs []string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return updateListsResultMsg{repoID: repoID, repoName: repoName, err: fmt.Errorf("not authenticated")}
		}
		ctx := context.Background()
		err := client.UpdateUserListsForItem(ctx, repoID, newListIDs)
		if err != nil {
			return updateListsResultMsg{repoID: repoID, repoName: repoName, err: err}
		}
		if memberships != nil {
			updateLocalMemberships(memberships, repoID, newListIDs)
			if store != nil {
				if err := store.SaveMemberships(memberships); err != nil {
					return updateListsResultMsg{
						repoID:   repoID,
						repoName: repoName,
						err:      fmt.Errorf("updated lists remotely but failed to save memberships locally: %w", err),
					}
				}
			}
		}
		return updateListsResultMsg{repoID: repoID, repoName: repoName}
	}
}

// createListCmd calls the GitHub API to create a new list.
func createListCmd(client *github.Client, name string) tea.Msg {
	ctx := context.Background()
	_, err := client.CreateList(ctx, name, "", false)
	if err != nil {
		return statusMsg{text: fmt.Sprintf("Create list failed: %s", err), isError: true}
	}
	return statusMsg{text: fmt.Sprintf("Created list %q – sync to refresh", name)}
}

// renameListCmd calls the GitHub API to rename a list.
func renameListCmd(client *github.Client, listID, name string) tea.Msg {
	ctx := context.Background()
	err := client.UpdateList(ctx, listID, name)
	if err != nil {
		return statusMsg{text: fmt.Sprintf("Rename list failed: %s", err), isError: true}
	}
	return statusMsg{text: fmt.Sprintf("Renamed list to %q – sync to refresh", name)}
}

// deleteListCmd calls the GitHub API to delete a list.
func deleteListCmd(client *github.Client, listID string) tea.Msg {
	ctx := context.Background()
	err := client.DeleteList(ctx, listID)
	if err != nil {
		return statusMsg{text: fmt.Sprintf("Delete list failed: %s", err), isError: true}
	}
	return statusMsg{text: "Deleted list – sync to refresh"}
}

// makeSyncCmd creates a tea.Cmd that runs a sync operation with progress reporting and a cancellable context.
func makeSyncCmd(token string, store *storage.Store, full bool, ctx context.Context, program *tea.Program) tea.Cmd {
	return func() tea.Msg {
		client := github.NewClient(token)
		engine := syncpkg.NewEngine(client, store)

		// Wire progress updates through tea.Program.Send so the UI updates in real time
		var onProgress syncpkg.SyncProgressFunc
		if program != nil {
			onProgress = func(p syncpkg.SyncProgress) {
				program.Send(syncProgressMsg{progress: p})
			}
		}

		var result *syncpkg.SyncResult
		var err error

		if full {
			result, err = engine.Full(ctx, onProgress)
		} else {
			result, err = engine.Delta(ctx, onProgress)
		}

		return syncCompleteMsg{result: result, err: err}
	}
}

// Run starts the Bubble Tea application.
func Run() error {
	store := storage.NewStore(storage.DefaultDir())
	authProv := auth.NewProvider(
		auth.WithKeyringBackend(auth.NewOSKeyring()),
	)
	m := NewModel(store, authProv)
	p := tea.NewProgram(&m, tea.WithAltScreen())
	m.SetProgram(p)
	_, err := p.Run()
	return err
}
