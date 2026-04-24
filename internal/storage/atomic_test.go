package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWrite(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, "test.json")

	data := []byte(`{"test": true}`)
	if err := AtomicWrite(path, data); err != nil {
		t.Fatalf("AtomicWrite: %v", err)
	}

	read, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(read) != string(data) {
		t.Errorf("got %q, want %q", string(read), string(data))
	}
}

func TestAtomicWrite_CreatesDir(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, "sub", "dir", "test.json")

	data := []byte(`{"nested": true}`)
	if err := AtomicWrite(path, data); err != nil {
		t.Fatalf("AtomicWrite: %v", err)
	}

	read, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(read) != string(data) {
		t.Errorf("got %q, want %q", string(read), string(data))
	}
}
