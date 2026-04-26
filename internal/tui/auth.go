package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Relequestual/astro-lab/internal/auth"
	"github.com/Relequestual/astro-lab/internal/github"
)

// authModel handles the authentication screen.
type authModel struct {
	textInput  textinput.Model
	spinner    spinner.Model
	err        string
	info       string
	validating bool
	useGHCLI   bool
	authProv   *auth.Provider
}

func newAuthModel(prov *auth.Provider) authModel {
	ti := textinput.New()
	ti.Placeholder = "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	ti.Prompt = "Token: "
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.CharLimit = 100
	ti.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return authModel{
		textInput: ti,
		spinner:   sp,
		authProv:  prov,
		info:      "Enter a GitHub personal access token, or press Tab to use gh CLI.",
	}
}

func (m authModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m authModel) Update(msg tea.Msg) (authModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.validating {
			return m, nil
		}
		switch msg.String() {
		case "tab":
			m.useGHCLI = !m.useGHCLI
			if m.useGHCLI {
				m.info = "Will authenticate via gh CLI. Press Enter to validate."
				m.textInput.Blur()
			} else {
				m.info = "Enter a GitHub personal access token, or press Tab to use gh CLI."
				m.textInput.Focus()
			}
			m.err = ""
			return m, nil
		case "ctrl+g":
			m.useGHCLI = true
			m.validating = true
			m.err = ""
			m.info = "Authenticating via gh CLI..."
			return m, validateGHCLICmd(m.authProv)
		case "enter":
			if m.useGHCLI {
				m.validating = true
				m.err = ""
				m.info = "Authenticating via gh CLI..."
				return m, validateGHCLICmd(m.authProv)
			}
			token := m.textInput.Value()
			if token == "" {
				m.err = "Token cannot be empty"
				return m, nil
			}
			m.validating = true
			m.err = ""
			m.info = "Validating token..."
			return m, validateTokenCmd(token, m.authProv)
		}

	case authValidatedMsg:
		m.validating = false
		if msg.err != nil {
			m.err = fmt.Sprintf("Authentication failed: %s", msg.err)
			m.info = ""
			return m, nil
		}
		// Success is handled by the parent model.
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if !m.useGHCLI && !m.validating {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m authModel) View() string {
	var s string

	s += titleStyle.Render("🔐 Authentication") + "\n\n"

	if m.useGHCLI {
		s += focusedStyle.Render("▸ gh CLI authentication") + "\n"
	} else {
		s += focusedStyle.Render("▸ Manual token entry") + "\n"
		s += m.textInput.View() + "\n"
	}
	s += "\n"

	if m.validating {
		s += m.spinner.View() + " " + m.info + "\n"
	} else if m.info != "" {
		s += mutedStyle.Render(m.info) + "\n"
	}

	if m.err != "" {
		s += "\n" + errorStyle.Render(m.err) + "\n"
	}

	s += "\n" + mutedStyle.Render("Tab: toggle method • Enter: validate • Ctrl+G: quick gh CLI")

	return s
}

// validateTokenCmd creates a command that validates a token against the GitHub API.
func validateTokenCmd(token string, prov *auth.Provider) tea.Cmd {
	return func() tea.Msg {
		client := github.NewClient(token)
		login, rl, err := client.ViewerLoginWithRateLimit(context.Background())
		if err != nil {
			return authValidatedMsg{err: err}
		}
		if prov != nil {
			_ = prov.StoreToken(token)
		}
		return authValidatedMsg{login: login, rateLimit: rl}
	}
}

// validateGHCLICmd creates a command that authenticates via gh CLI.
func validateGHCLICmd(prov *auth.Provider) tea.Cmd {
	return func() tea.Msg {
		if prov == nil {
			return authValidatedMsg{err: fmt.Errorf("no auth provider configured")}
		}
		token, _, err := prov.Resolve()
		if err != nil {
			return authValidatedMsg{err: err}
		}
		client := github.NewClient(token)
		login, rl, err := client.ViewerLoginWithRateLimit(context.Background())
		if err != nil {
			return authValidatedMsg{err: err}
		}
		return authValidatedMsg{login: login, rateLimit: rl}
	}
}
