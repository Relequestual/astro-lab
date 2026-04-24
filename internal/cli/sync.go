package cli

import (
	"context"
	"fmt"

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
	if syncFull {
		if !jsonOutput {
			fmt.Println("Performing full sync...")
		}
		result, err = engine.Full(context.Background())
	} else {
		if !jsonOutput {
			fmt.Println("Performing delta sync...")
		}
		result, err = engine.Delta(context.Background())
	}

	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	if outputJSON(result) {
		return nil
	}

	fmt.Println(result.String())
	return nil
}
