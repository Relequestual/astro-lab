package cli

import (
	"bytes"
	"testing"
)

func TestAuthCommands_Registered(t *testing.T) {
	// Verify all auth subcommands are registered
	cmds := authCmd.Commands()
	names := make(map[string]bool)
	for _, c := range cmds {
		names[c.Name()] = true
	}

	expected := []string{"login", "status", "logout"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing auth subcommand: %s", name)
		}
	}
}

func TestRootCommand_Flags(t *testing.T) {
	f := rootCmd.PersistentFlags()

	if f.Lookup("json") == nil {
		t.Error("missing --json flag")
	}
	if f.Lookup("token") == nil {
		t.Error("missing --token flag")
	}
}

func TestOutputJSON(t *testing.T) {
	// Test with json mode off
	jsonOutput = false
	if outputJSON("test") {
		t.Error("expected false when json mode is off")
	}

	// Test with json mode on
	jsonOutput = true
	defer func() { jsonOutput = false }()

	// Capture would need redirecting stdout, just verify it returns true
	old := jsonOutput
	jsonOutput = true
	result := outputJSON(map[string]string{"test": "value"})
	jsonOutput = old
	if !result {
		t.Error("expected true when json mode is on")
	}
}

func TestSyncCommand_Registered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "sync" {
			found = true
			// Check --full flag exists
			if c.Flags().Lookup("full") == nil {
				t.Error("missing --full flag on sync command")
			}
			break
		}
	}
	if !found {
		t.Error("sync command not registered")
	}
}

func TestMoveCommand_Registered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "move" {
			found = true
			if c.Flags().Lookup("repo") == nil {
				t.Error("missing --repo flag")
			}
			if c.Flags().Lookup("lists") == nil {
				t.Error("missing --lists flag")
			}
			if c.Flags().Lookup("apply") == nil {
				t.Error("missing --apply flag")
			}
			if c.Flags().Lookup("force") == nil {
				t.Error("missing --force flag")
			}
			break
		}
	}
	if !found {
		t.Error("move command not registered")
	}
}

func TestListsCommand_Registered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "lists" {
			found = true
			break
		}
	}
	if !found {
		t.Error("lists command not registered")
	}
}

func TestStarsCommand_Registered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "stars" {
			found = true
			if c.Flags().Lookup("limit") == nil {
				t.Error("missing --limit flag")
			}
			if c.Flags().Lookup("since") == nil {
				t.Error("missing --since flag")
			}
			break
		}
	}
	if !found {
		t.Error("stars command not registered")
	}
}

func TestRootCommand_Execute_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
