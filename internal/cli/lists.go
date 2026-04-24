package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Relequestual/astro-lab/internal/github"
	"github.com/Relequestual/astro-lab/internal/models"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

var listsCmd = &cobra.Command{
	Use:   "lists",
	Short: "Show star lists",
	RunE:  runLists,
}

func init() {
	rootCmd.AddCommand(listsCmd)
}

func runLists(cmd *cobra.Command, args []string) error {
	token, err := resolveToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	client := github.NewClient(token)
	var lists []models.StarList
	var fetchErr error
	action := func() {
		lists, fetchErr = client.FetchLists(cmd.Context())
	}
	if err := spinner.New().Title("Fetching lists...").Action(action).Run(); err != nil {
		return fmt.Errorf("spinner: %w", err)
	}
	if fetchErr != nil {
		return fmt.Errorf("fetching lists: %w", fetchErr)
	}

	if outputJSON(lists) {
		return nil
	}

	if len(lists) == 0 {
		fmt.Println("No star lists found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "NAME\tID\tITEMS\tPRIVATE\n")
	for _, l := range lists {
		privacy := "public"
		if l.IsPrivate {
			privacy = "private"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", l.Name, l.ID, l.ItemCount, privacy)
	}
	w.Flush()
	return nil
}
