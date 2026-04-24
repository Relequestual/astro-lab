package auth

import (
	"testing"

	"github.com/Relequestual/astro-lab/internal/models"
)

// mockKeyring is a test keyring backend
type mockKeyring struct {
	store     map[string]string
	available bool
}

func newMockKeyring(available bool) *mockKeyring {
	return &mockKeyring{
		store:     make(map[string]string),
		available: available,
	}
}

func (m *mockKeyring) Get(service, key string) (string, error) {
	v, ok := m.store[service+":"+key]
	if !ok {
		return "", ErrNoAuth
	}
	return v, nil
}

func (m *mockKeyring) Set(service, key, value string) error {
	m.store[service+":"+key] = value
	return nil
}

func (m *mockKeyring) Delete(service, key string) error {
	delete(m.store, service+":"+key)
	return nil
}

func (m *mockKeyring) Available() bool {
	return m.available
}

// mockGH is a test GH backend
type mockGH struct {
	token string
	err   error
}

func (m *mockGH) Token() (string, error) {
	return m.token, m.err
}

func TestProviderResolve_ExplicitToken(t *testing.T) {
	p := NewProvider(WithExplicitToken("test-token"))
	token, provider, err := p.Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "test-token" {
		t.Errorf("expected test-token, got %s", token)
	}
	if provider != models.AuthProviderToken {
		t.Errorf("expected token provider, got %s", provider)
	}
}

func TestProviderResolve_EnvToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "env-token")
	p := NewProvider()
	token, provider, err := p.Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "env-token" {
		t.Errorf("expected env-token, got %s", token)
	}
	if provider != models.AuthProviderEnv {
		t.Errorf("expected env provider, got %s", provider)
	}
}

func TestProviderResolve_Keyring(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	kr := newMockKeyring(true)
	kr.store[keyringService+":"+keyringKey] = "keyring-token"

	p := NewProvider(
		WithKeyringBackend(kr),
	)
	token, provider, err := p.Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "keyring-token" {
		t.Errorf("expected keyring-token, got %s", token)
	}
	if provider != models.AuthProviderKeyring {
		t.Errorf("expected keyring provider, got %s", provider)
	}
}

func TestProviderResolve_GHFallback(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	gh := &mockGH{token: "gh-token"}
	kr := newMockKeyring(false)

	p := NewProvider(
		WithKeyringBackend(kr),
		WithGHBackend(gh),
	)
	token, provider, err := p.Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "gh-token" {
		t.Errorf("expected gh-token, got %s", token)
	}
	if provider != models.AuthProviderGH {
		t.Errorf("expected gh provider, got %s", provider)
	}
}

func TestProviderResolve_NoAuth(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	gh := &mockGH{err: ErrNoAuth}
	kr := newMockKeyring(false)

	p := NewProvider(
		WithKeyringBackend(kr),
		WithGHBackend(gh),
	)
	_, _, err := p.Resolve()
	if err != ErrNoAuth {
		t.Errorf("expected ErrNoAuth, got %v", err)
	}
}

func TestProviderResolve_Precedence(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "env-token")
	kr := newMockKeyring(true)
	kr.store[keyringService+":"+keyringKey] = "keyring-token"
	gh := &mockGH{token: "gh-token"}

	// Explicit token should win
	p := NewProvider(
		WithExplicitToken("explicit-token"),
		WithKeyringBackend(kr),
		WithGHBackend(gh),
	)
	token, provider, _ := p.Resolve()
	if token != "explicit-token" || provider != models.AuthProviderToken {
		t.Errorf("explicit token should have highest precedence")
	}
}

func TestProviderStoreToken_KeyringAvailable(t *testing.T) {
	kr := newMockKeyring(true)
	p := NewProvider(
		WithKeyringBackend(kr),
		WithStore(models.AuthStoreKeyring),
	)

	if err := p.StoreToken("my-token"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, err := kr.Get(keyringService, keyringKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stored != "my-token" {
		t.Errorf("expected my-token, got %s", stored)
	}
}

func TestProviderStoreToken_KeyringUnavailable(t *testing.T) {
	kr := newMockKeyring(false)
	p := NewProvider(
		WithKeyringBackend(kr),
		WithStore(models.AuthStoreKeyring),
	)

	err := p.StoreToken("my-token")
	if err != ErrKeyringUnavailable {
		t.Errorf("expected ErrKeyringUnavailable, got %v", err)
	}
}

func TestProviderStoreToken_NoPersist(t *testing.T) {
	p := NewProvider(
		WithStore(models.AuthStoreNone),
	)

	if err := p.StoreToken("my-token"); err != nil {
		t.Errorf("no-persist store should not error: %v", err)
	}
}

func TestProviderClearToken(t *testing.T) {
	kr := newMockKeyring(true)
	kr.store[keyringService+":"+keyringKey] = "my-token"

	p := NewProvider(WithKeyringBackend(kr))
	if err := p.ClearToken(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := kr.Get(keyringService, keyringKey)
	if err == nil {
		t.Error("expected token to be deleted")
	}
}
