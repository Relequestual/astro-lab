package auth

import (
	"github.com/zalando/go-keyring"
)

type OSKeyring struct{}

func NewOSKeyring() *OSKeyring {
	return &OSKeyring{}
}

func (k *OSKeyring) Available() bool {
	// Test availability by attempting a get for a non-existent key
	_, err := keyring.Get(keyringService, "availability-check")
	// If we get ErrNotFound, keyring is available but key doesn't exist
	if err == keyring.ErrNotFound {
		return true
	}
	// Some systems return nil if the key doesn't exist
	if err == nil {
		return true
	}
	return false
}

func (k *OSKeyring) Get(service, key string) (string, error) {
	return keyring.Get(service, key)
}

func (k *OSKeyring) Set(service, key, value string) error {
	return keyring.Set(service, key, value)
}

func (k *OSKeyring) Delete(service, key string) error {
	return keyring.Delete(service, key)
}
