package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// searchModel is a reusable inline search component.
type searchModel struct {
	input textinput.Model
	query string
}

func newSearchModel(placeholder string) searchModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = "/ "
	ti.CharLimit = 120
	return searchModel{input: ti}
}

// Focus activates the search input.
func (s *searchModel) Focus() {
	s.input.Focus()
}

// Reset clears the query and blurs the input.
func (s *searchModel) Reset() {
	s.input.SetValue("")
	s.query = ""
	s.input.Blur()
}

// Query returns the current search query.
func (s *searchModel) Query() string {
	return s.query
}

func (s searchModel) Init() tea.Cmd {
	return nil
}

func (s searchModel) Update(msg tea.Msg) (searchModel, tea.Cmd) {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	s.query = s.input.Value()
	return s, cmd
}

func (s searchModel) View() string {
	return s.input.View()
}
