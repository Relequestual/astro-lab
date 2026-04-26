package tui

import "testing"

func TestUndoPushPop(t *testing.T) {
	s := NewUndoStack(10)
	s.Push(UndoEntry{Description: "first", RepoID: "r1"})
	s.Push(UndoEntry{Description: "second", RepoID: "r2"})

	if s.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", s.Len())
	}

	e, ok := s.Pop()
	if !ok {
		t.Fatal("Pop() returned false, want true")
	}
	if e.Description != "second" {
		t.Errorf("Pop().Description = %q, want %q", e.Description, "second")
	}

	e, ok = s.Pop()
	if !ok {
		t.Fatal("Pop() returned false, want true")
	}
	if e.Description != "first" {
		t.Errorf("Pop().Description = %q, want %q", e.Description, "first")
	}

	_, ok = s.Pop()
	if ok {
		t.Error("Pop() on empty stack returned true, want false")
	}
}

func TestUndoEviction(t *testing.T) {
	s := NewUndoStack(3)
	s.Push(UndoEntry{Description: "a"})
	s.Push(UndoEntry{Description: "b"})
	s.Push(UndoEntry{Description: "c"})
	s.Push(UndoEntry{Description: "d"})

	if s.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", s.Len())
	}

	// Oldest ("a") should have been evicted
	e, _ := s.Pop()
	if e.Description != "d" {
		t.Errorf("Pop() = %q, want %q", e.Description, "d")
	}
	e, _ = s.Pop()
	if e.Description != "c" {
		t.Errorf("Pop() = %q, want %q", e.Description, "c")
	}
	e, _ = s.Pop()
	if e.Description != "b" {
		t.Errorf("Pop() = %q, want %q", e.Description, "b")
	}
}

func TestUndoPeek(t *testing.T) {
	s := NewUndoStack(5)

	_, ok := s.Peek()
	if ok {
		t.Error("Peek() on empty stack returned true, want false")
	}

	s.Push(UndoEntry{Description: "first"})
	s.Push(UndoEntry{Description: "second"})

	e, ok := s.Peek()
	if !ok {
		t.Fatal("Peek() returned false, want true")
	}
	if e.Description != "second" {
		t.Errorf("Peek().Description = %q, want %q", e.Description, "second")
	}
	// Peek should not remove the entry
	if s.Len() != 2 {
		t.Errorf("Len() after Peek = %d, want 2", s.Len())
	}
}

func TestUndoClear(t *testing.T) {
	s := NewUndoStack(10)
	s.Push(UndoEntry{Description: "a"})
	s.Push(UndoEntry{Description: "b"})
	s.Clear()

	if s.Len() != 0 {
		t.Errorf("Len() after Clear = %d, want 0", s.Len())
	}
	_, ok := s.Pop()
	if ok {
		t.Error("Pop() after Clear returned true, want false")
	}
}

func TestUndoString(t *testing.T) {
	s := NewUndoStack(5)

	got := s.String()
	if got != "undo stack empty" {
		t.Errorf("String() on empty = %q, want %q", got, "undo stack empty")
	}

	s.Push(UndoEntry{Description: "move repo to list-A"})
	got = s.String()
	if got != "undo (1): move repo to list-A" {
		t.Errorf("String() = %q, want %q", got, "undo (1): move repo to list-A")
	}

	s.Push(UndoEntry{Description: "second action"})
	got = s.String()
	if got != "undo (2): second action" {
		t.Errorf("String() = %q, want %q", got, "undo (2): second action")
	}
}

func TestUndoMinMaxSize(t *testing.T) {
	s := NewUndoStack(0) // should be clamped to 1
	s.Push(UndoEntry{Description: "a"})
	s.Push(UndoEntry{Description: "b"})
	if s.Len() != 1 {
		t.Errorf("Len() with maxSize=0 = %d, want 1", s.Len())
	}
	e, _ := s.Pop()
	if e.Description != "b" {
		t.Errorf("Pop().Description = %q, want %q", e.Description, "b")
	}
}
