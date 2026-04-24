package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	tokenFlag  string
)

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "astlab",
	Short: "Astrometrics Lab - Manage GitHub stars and star lists",
	Long:  "A CLI and TUI for managing GitHub stars and star lists safely.",
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().StringVar(&tokenFlag, "token", "", "GitHub token (overrides all other auth sources)")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// outputJSON outputs data as JSON if --json flag is set, returns true if it handled output
func outputJSON(v interface{}) bool {
	if !jsonOutput {
		return false
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
	}
	return true
}
