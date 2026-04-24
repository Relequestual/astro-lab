package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Relequestual/astro-lab/internal/github"
	"github.com/Relequestual/astro-lab/internal/storage"
	gosync "github.com/Relequestual/astro-lab/internal/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync stars and lists from GitHub",
	RunE:  runSync,
}

var syncFull bool

func init() {
	syncCmd.Flags().BoolVar(&syncFull, "full", false, "Perform full reconciliation sync")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	token, err := resolveToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	client := github.NewClient(token)
	store := storage.NewStore(storage.DefaultDir())
	engine := gosync.NewEngine(client, store)

	var result *gosync.SyncResult
	var syncErr error

	title := "Performing delta sync..."
	if syncFull {
		title = "Performing full sync..."
	}

	err = runWithProgress(title, func(p *tea.Program) error {
		onProgress := func(sp gosync.SyncProgress) {
			var text string
			switch sp.Phase {
			case gosync.PhaseStars:
				text = fmt.Sprintf("Fetching stars... %d/%d", sp.Fetched, sp.Total)
			case gosync.PhaseLists:
				text = fmt.Sprintf("Fetching lists... %d/%d", sp.Fetched, sp.Total)
			case gosync.PhaseMemberships:
				text = fmt.Sprintf("Fetching list items... %d/%d lists", sp.Fetched, sp.Total)
			}
			p.Send(progressUpdate{text: text})
		}

		if syncFull {
			result, syncErr = engine.Full(cmd.Context(), onProgress)
		} else {
			result, syncErr = engine.Delta(cmd.Context(), onProgress)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	if syncErr != nil {
		return fmt.Errorf("sync failed: %w", syncErr)
	}

	if outputJSON(result) {
		return nil
	}

	fmt.Println(result.String())
	return nil
}
