package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/Relequestual/astro-lab/internal/github"
	"github.com/spf13/cobra"
)

var starsCmd = &cobra.Command{
	Use:   "stars",
	Short: "Show starred repositories",
	RunE:  runStars,
}

var (
	starsLimit int
	starsSince string
)

func init() {
	starsCmd.Flags().IntVar(&starsLimit, "limit", 50, "Maximum number of stars to show")
	starsCmd.Flags().StringVar(&starsSince, "since", "", "Show stars since date (YYYY-MM-DD)")
	rootCmd.AddCommand(starsCmd)
}

func runStars(cmd *cobra.Command, args []string) error {
	token, err := resolveToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	var since time.Time
	if starsSince != "" {
		since, err = time.Parse("2006-01-02", starsSince)
		if err != nil {
			return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
		}
	}

	client := github.NewClient(token)
	stars, err := client.FetchStarredRepos(cmd.Context(), since)
	if err != nil {
		return fmt.Errorf("fetching stars: %w", err)
	}

	// Apply limit
	if starsLimit > 0 && len(stars) > starsLimit {
		stars = stars[:starsLimit]
	}

	if outputJSON(stars) {
		return nil
	}

	if len(stars) == 0 {
		fmt.Println("No starred repositories found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "REPOSITORY\tSTARRED AT\n")
	for _, s := range stars {
		fmt.Fprintf(w, "%s\t%s\n", s.NameWithOwner, s.StarredAt.Format("2006-01-02"))
	}
	w.Flush()
	return nil
}
