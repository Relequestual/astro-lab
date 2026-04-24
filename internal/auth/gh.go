package auth

import (
	"os/exec"
	"strings"
)

type defaultGHBackend struct{}

func (g *defaultGHBackend) Token() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// DefaultGHBackendExported is an exported wrapper around defaultGHBackend
// for use by the CLI auth login command.
type DefaultGHBackendExported struct{}

func (g *DefaultGHBackendExported) Token() (string, error) {
	return (&defaultGHBackend{}).Token()
}
