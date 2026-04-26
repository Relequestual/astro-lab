package tui

import "fmt"

// UndoEntry records a reversible action.
type UndoEntry struct {
	Description string
	RepoID      string
	RepoName    string
	PreviousIDs []string
}

// UndoStack is a bounded LIFO stack of undo entries.
type UndoStack struct {
	entries []UndoEntry
	maxSize int
}

// NewUndoStack creates an undo stack with the given capacity.
func NewUndoStack(maxSize int) *UndoStack {
	if maxSize < 1 {
		maxSize = 1
	}
	return &UndoStack{maxSize: maxSize}
}

// Push adds an entry, evicting the oldest if at capacity.
func (s *UndoStack) Push(e UndoEntry) {
	if len(s.entries) >= s.maxSize {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, e)
}

// Pop removes and returns the most recent entry. Returns false if empty.
func (s *UndoStack) Pop() (UndoEntry, bool) {
	if len(s.entries) == 0 {
		return UndoEntry{}, false
	}
	e := s.entries[len(s.entries)-1]
	s.entries = s.entries[:len(s.entries)-1]
	return e, true
}

// Len returns the number of entries.
func (s *UndoStack) Len() int {
	return len(s.entries)
}

// Peek returns the most recent entry without removing it. Returns false if empty.
func (s *UndoStack) Peek() (UndoEntry, bool) {
	if len(s.entries) == 0 {
		return UndoEntry{}, false
	}
	return s.entries[len(s.entries)-1], true
}

// Clear removes all entries.
func (s *UndoStack) Clear() {
	s.entries = nil
}

// String returns a human-readable summary.
func (s *UndoStack) String() string {
	if len(s.entries) == 0 {
		return "undo stack empty"
	}
	top := s.entries[len(s.entries)-1]
	return fmt.Sprintf("undo (%d): %s", len(s.entries), top.Description)
}
