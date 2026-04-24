package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/Relequestual/astro-lab/internal/auth"
	"github.com/Relequestual/astro-lab/internal/github"
	"github.com/Relequestual/astro-lab/internal/models"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
	// Validate flag values
	if authMethod != "token" && authMethod != "gh" {
		return fmt.Errorf("invalid --method %q: must be \"token\" or \"gh\"", authMethod)
	}
	if authStore != "keyring" && authStore != "none" {
		return fmt.Errorf("invalid --store %q: must be \"keyring\" or \"none\"", authStore)
	}

	store := models.AuthStoreKeyring
	if authStore == "none" {
		store = models.AuthStoreNone
	}

	if authMethod == "gh" {
		// For gh method, resolve token directly from gh CLI
		ghBackend := &auth.DefaultGHBackendExported{}
		token, err := ghBackend.Token()
		if err != nil {
			return fmt.Errorf("failed to get token from gh CLI: %w", err)
		}

		// Validate token
		client := github.NewClient(token)
		var login string
		var validateErr error
		validateAction := func() {
			login, validateErr = client.ViewerLogin(cmd.Context())
		}
		if err := spinner.New().Title("Validating token...").Action(validateAction).Run(); err != nil {
			return fmt.Errorf("spinner: %w", err)
		}
		if validateErr != nil {
			return fmt.Errorf("failed to validate token: %w", validateErr)
		}

		// Store token if requested
		if store == models.AuthStoreKeyring {
			kb := auth.NewOSKeyring()
			provider := auth.NewProvider(
				auth.WithStore(store),
				auth.WithKeyringBackend(kb),
				auth.WithExplicitToken(token),
			)
			if err := provider.StoreToken(token); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not store token in keyring: %v\n", err)
			} else {
				fmt.Println("Token stored in keyring.")
			}
		}

		fmt.Printf("✓ Authenticated via gh CLI as %s\n", login)
		return nil
	}

	// Token method: read from stdin
	var tokenInput string
	if term.IsTerminal(int(os.Stdin.Fd())) {
		// Interactive TTY: use no-echo input
		fmt.Print("Enter GitHub token: ")
		tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println() // newline after hidden input
		if err != nil {
			return fmt.Errorf("reading token: %w", err)
		}
		tokenInput = strings.TrimSpace(string(tokenBytes))
		if tokenInput == "" {
			return fmt.Errorf("empty token provided")
		}
	} else {
		// Piped/redirected stdin: read line
		var err error
		tokenInput, err = auth.ReadTokenFromInput(os.Stdin)
		if err != nil {
			return err
		}
	}

	// Validate token
	client := github.NewClient(tokenInput)
	var login string
	var validateErr error
	validateAction := func() {
		login, validateErr = client.ViewerLogin(cmd.Context())
	}
	if err := spinner.New().Title("Validating token...").Action(validateAction).Run(); err != nil {
		return fmt.Errorf("spinner: %w", err)
	}
	if validateErr != nil {
		return fmt.Errorf("invalid token: %w", validateErr)
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
	var login string
	var validateErr error
	validateAction := func() {
		login, validateErr = client.ViewerLogin(cmd.Context())
	}
	if err := spinner.New().Title("Checking auth...").Action(validateAction).Run(); err != nil {
		return fmt.Errorf("spinner: %w", err)
	}
	if validateErr != nil {
		status := models.AuthStatus{
			Provider:      authProv,
			Authenticated: false,
		}
		if outputJSON(status) {
			return nil
		}
		fmt.Printf("Token found via %s but validation failed: %s\n", authProv, auth.RedactSecrets(validateErr.Error()))
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
