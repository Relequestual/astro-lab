package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/Relequestual/astro-lab/internal/models"
)

var (
	ErrNoAuth             = errors.New("no authentication source available")
	ErrKeyringUnavailable = errors.New("keyring is not available; use --store none")
)

// Provider resolves a GitHub token from various sources
type Provider struct {
	explicitToken  string
	store          models.AuthStore
	keyringAvail   bool
	keyringBackend KeyringBackend
	ghBackend      GHBackend
}

type KeyringBackend interface {
	Get(service, key string) (string, error)
	Set(service, key, value string) error
	Delete(service, key string) error
	Available() bool
}

type GHBackend interface {
	Token() (string, error)
}

const (
	keyringService = "astro-lab"
	keyringKey     = "github-token"
)

func NewProvider(opts ...ProviderOption) *Provider {
	p := &Provider{
		store: models.AuthStoreNone,
	}
	for _, o := range opts {
		o(p)
	}
	if p.keyringBackend != nil {
		p.keyringAvail = p.keyringBackend.Available()
	}
	if p.ghBackend == nil {
		p.ghBackend = &defaultGHBackend{}
	}
	return p
}

type ProviderOption func(*Provider)

func WithExplicitToken(t string) ProviderOption {
	return func(p *Provider) { p.explicitToken = t }
}

func WithStore(s models.AuthStore) ProviderOption {
	return func(p *Provider) { p.store = s }
}

func WithKeyringBackend(kb KeyringBackend) ProviderOption {
	return func(p *Provider) { p.keyringBackend = kb }
}

func WithGHBackend(gb GHBackend) ProviderOption {
	return func(p *Provider) { p.ghBackend = gb }
}

// Resolve returns a token following the precedence chain:
// 1. Explicit token (flag)
// 2. GITHUB_TOKEN env
// 3. Keyring stored token
// 4. gh auth token fallback
func (p *Provider) Resolve() (string, models.AuthProvider, error) {
	// 1. Explicit token
	if p.explicitToken != "" {
		return p.explicitToken, models.AuthProviderToken, nil
	}

	// 2. Environment variable
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t, models.AuthProviderEnv, nil
	}

	// 3. Keyring
	if p.keyringAvail && p.keyringBackend != nil {
		t, err := p.keyringBackend.Get(keyringService, keyringKey)
		if err == nil && t != "" {
			return t, models.AuthProviderKeyring, nil
		}
	}

	// 4. gh fallback
	if p.ghBackend != nil {
		t, err := p.ghBackend.Token()
		if err == nil && t != "" {
			return t, models.AuthProviderGH, nil
		}
	}

	return "", "", ErrNoAuth
}

// StoreToken stores a token in the keyring
func (p *Provider) StoreToken(token string) error {
	if p.store == models.AuthStoreNone {
		return nil // no-persist mode
	}
	if !p.keyringAvail || p.keyringBackend == nil {
		return ErrKeyringUnavailable
	}
	return p.keyringBackend.Set(keyringService, keyringKey, token)
}

// ClearToken removes the stored token
func (p *Provider) ClearToken() error {
	if p.keyringBackend != nil && p.keyringAvail {
		return p.keyringBackend.Delete(keyringService, keyringKey)
	}
	return nil
}

// ProviderInfo returns a display-safe description
func (p *Provider) ProviderInfo() string {
	_, provider, err := p.Resolve()
	if err != nil {
		return fmt.Sprintf("not authenticated: %s", RedactSecrets(err.Error()))
	}
	return fmt.Sprintf("authenticated via %s", provider)
}
