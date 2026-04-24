package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Relequestual/astro-lab/internal/github"
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
	lists, err := client.FetchLists(context.Background())
	if err != nil {
		return fmt.Errorf("fetching lists: %w", err)
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
