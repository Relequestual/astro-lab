package cli

import (
	"fmt"
	"strings"

	"github.com/Relequestual/astro-lab/internal/github"
	"github.com/Relequestual/astro-lab/internal/mutation"
	"github.com/Relequestual/astro-lab/internal/storage"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move a repository between star lists",
	Long:  "Update star list memberships for a repository. Default is dry run; use --apply to execute.",
	RunE:  runMove,
}

var (
	moveRepo    string
	moveToLists []string
	moveApply   bool
	moveForce   bool
)

func init() {
	moveCmd.Flags().StringVar(&moveRepo, "repo", "", "Repository ID or name (required)")
	moveCmd.Flags().StringSliceVar(&moveToLists, "lists", nil, "Target list IDs (comma-separated)")
	moveCmd.Flags().BoolVar(&moveApply, "apply", false, "Execute the changes (default is dry run)")
	moveCmd.Flags().BoolVar(&moveForce, "force", false, "Skip removal confirmation")
	_ = moveCmd.MarkFlagRequired("repo")
	_ = moveCmd.MarkFlagRequired("lists")
	rootCmd.AddCommand(moveCmd)
}

func runMove(cmd *cobra.Command, args []string) error {
	token, err := resolveToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	client := github.NewClient(token)
	store := storage.NewStore(storage.DefaultDir())

	// Resolve repo ID from name if needed
	repoID, repoName, err := resolveRepoID(store, moveRepo)
	if err != nil {
		return fmt.Errorf("resolving repository: %w", err)
	}

	engine := mutation.NewMoveEngine(client, store)

	// Plan the move
	var result *mutation.MoveResult
	var planErr error
	planAction := func() {
		result, planErr = engine.Plan(cmd.Context(), repoID, repoName, moveToLists)
	}
	if err := spinner.New().Title("Planning move...").Action(planAction).Run(); err != nil {
		return fmt.Errorf("spinner: %w", err)
	}
	if planErr != nil {
		return fmt.Errorf("planning move: %w", planErr)
	}

	if outputJSON(result) {
		return nil
	}

	// Display diff
	displayMoveDiff(result, store)

	if !moveApply {
		fmt.Println("\nDry run - no changes made. Use --apply to execute.")
		return nil
	}

	// Check for removals and confirm
	if len(result.Diff.Removed) > 0 && !moveForce {
		fmt.Printf("\n⚠ This will remove the repository from %d list(s). Continue? [y/N] ", len(result.Diff.Removed))
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Apply
	var applyErr error
	applyAction := func() {
		result, applyErr = engine.Apply(cmd.Context(), repoID, repoName, moveToLists)
	}
	if err := spinner.New().Title("Applying changes...").Action(applyAction).Run(); err != nil {
		return fmt.Errorf("spinner: %w", err)
	}
	if applyErr != nil {
		return fmt.Errorf("applying move: %w", applyErr)
	}

	fmt.Println("✓ Changes applied successfully.")
	return nil
}

func resolveRepoID(store *storage.Store, input string) (string, string, error) {
	stars, err := store.LoadStars()
	if err != nil {
		return input, input, nil // Fall back to treating input as ID
	}

	// Check if input is a repo ID
	if repo, ok := stars.ByRepoID[input]; ok {
		return repo.ID, repo.NameWithOwner, nil
	}

	// Check if input matches a nameWithOwner
	for _, repo := range stars.ByRepoID {
		if strings.EqualFold(repo.NameWithOwner, input) {
			return repo.ID, repo.NameWithOwner, nil
		}
	}

	return input, input, nil // Fall back to treating input as ID
}

func displayMoveDiff(result *mutation.MoveResult, store *storage.Store) {
	lists, _ := store.LoadLists()
	nameForID := func(id string) string {
		if lists != nil {
			if l, ok := lists.ByListID[id]; ok {
				return l.Name
			}
		}
		return id
	}

	fmt.Printf("Repository: %s\n\n", result.RepoName)

	if len(result.Diff.Before) > 0 {
		fmt.Println("Current lists:")
		for _, id := range result.Diff.Before {
			fmt.Printf("  • %s\n", nameForID(id))
		}
	} else {
		fmt.Println("Current lists: (none)")
	}

	fmt.Println()

	if len(result.Diff.After) > 0 {
		fmt.Println("Target lists:")
		for _, id := range result.Diff.After {
			fmt.Printf("  • %s\n", nameForID(id))
		}
	} else {
		fmt.Println("Target lists: (none)")
	}

	if len(result.Diff.Added) > 0 {
		fmt.Println("\nAdding to:")
		for _, id := range result.Diff.Added {
			fmt.Printf("  + %s\n", nameForID(id))
		}
	}

	if len(result.Diff.Removed) > 0 {
		fmt.Println("\nRemoving from:")
		for _, id := range result.Diff.Removed {
			fmt.Printf("  - %s\n", nameForID(id))
		}
	}

	if len(result.Diff.Added) == 0 && len(result.Diff.Removed) == 0 {
		fmt.Println("\nNo changes needed.")
	}
}
