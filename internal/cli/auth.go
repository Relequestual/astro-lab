package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Relequestual/astro-lab/internal/auth"
	"github.com/Relequestual/astro-lab/internal/github"
	"github.com/Relequestual/astro-lab/internal/models"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub",
	RunE:  runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	RunE:  runAuthStatus,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored authentication",
	RunE:  runAuthLogout,
}

var (
	authMethod string
	authStore  string
)

func init() {
	authLoginCmd.Flags().StringVar(&authMethod, "method", "token", "Auth method: token or gh")
	authLoginCmd.Flags().StringVar(&authStore, "store", "keyring", "Token storage: keyring or none")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	store := models.AuthStoreKeyring
	if authStore == "none" {
		store = models.AuthStoreNone
	}

	var opts []auth.ProviderOption
	opts = append(opts, auth.WithStore(store))

	if authMethod == "gh" {
		// For gh method, resolve token from gh CLI
		opts = append(opts, auth.WithGHBackend(&auth.DefaultGHBackendExported{}))
		provider := auth.NewProvider(opts...)
		token, authProv, err := provider.Resolve()
		if err != nil {
			return fmt.Errorf("failed to authenticate via gh: %w", err)
		}

		// Validate token
		client := github.NewClient(token)
		login, err := client.ViewerLogin(context.Background())
		if err != nil {
			return fmt.Errorf("failed to validate token: %w", err)
		}

		fmt.Printf("✓ Authenticated via %s as %s\n", authProv, login)
		return nil
	}

	// Token method: read from stdin or prompt
	fmt.Print("Enter GitHub token: ")
	reader := bufio.NewReader(os.Stdin)
	tokenInput, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading token: %w", err)
	}
	tokenInput = strings.TrimSpace(tokenInput)

	if tokenInput == "" {
		return fmt.Errorf("empty token provided")
	}

	// Validate token
	client := github.NewClient(tokenInput)
	login, err := client.ViewerLogin(context.Background())
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Store token if requested
	if store == models.AuthStoreKeyring {
		kb := auth.NewOSKeyring()
		provider := auth.NewProvider(
			auth.WithStore(store),
			auth.WithKeyringBackend(kb),
			auth.WithExplicitToken(tokenInput),
		)
		if err := provider.StoreToken(tokenInput); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not store token in keyring: %v\n", err)
			fmt.Println("Token will be used for this session only.")
		} else {
			fmt.Println("Token stored in keyring.")
		}
	}

	fmt.Printf("✓ Authenticated as %s\n", login)
	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	provider := resolveAuthProvider()
	token, authProv, err := provider.Resolve()
	if err != nil {
		status := models.AuthStatus{Authenticated: false}
		if outputJSON(status) {
			return nil
		}
		fmt.Println("Not authenticated.")
		fmt.Printf("Error: %s\n", auth.RedactSecrets(err.Error()))
		return nil
	}

	client := github.NewClient(token)
	login, err := client.ViewerLogin(context.Background())
	if err != nil {
		status := models.AuthStatus{
			Provider:      authProv,
			Authenticated: false,
		}
		if outputJSON(status) {
			return nil
		}
		fmt.Printf("Token found via %s but validation failed: %s\n", authProv, auth.RedactSecrets(err.Error()))
		return nil
	}

	status := models.AuthStatus{
		Provider:      authProv,
		Login:         login,
		Authenticated: true,
	}

	if outputJSON(status) {
		return nil
	}

	fmt.Printf("✓ Authenticated via %s as %s\n", authProv, login)
	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	kb := auth.NewOSKeyring()
	provider := auth.NewProvider(
		auth.WithKeyringBackend(kb),
		auth.WithStore(models.AuthStoreKeyring),
	)
	if err := provider.ClearToken(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
	fmt.Println("Logged out.")
	return nil
}

func resolveAuthProvider() *auth.Provider {
	var opts []auth.ProviderOption

	if tokenFlag != "" {
		opts = append(opts, auth.WithExplicitToken(tokenFlag))
	}

	kb := auth.NewOSKeyring()
	opts = append(opts, auth.WithKeyringBackend(kb))

	return auth.NewProvider(opts...)
}

func resolveToken() (string, error) {
	provider := resolveAuthProvider()
	token, _, err := provider.Resolve()
	return token, err
}
