package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/activtrak-mfinlayson/atadmin/internal/api"
	"github.com/activtrak-mfinlayson/atadmin/internal/config"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)

const tokenURL = "https://app.activtrak.com/#/app/integrations/apikeys"

func newAuthCmd() *cobra.Command {
	auth := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  `Commands for authenticating with the ActivTrak API.`,
	}
	auth.AddCommand(newAuthLoginCmd())
	return auth
}

func newAuthLoginCmd() *cobra.Command {
	var profileFlag string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the ActivTrak API",
		Long: `Open your browser to the API token generation page, paste your token,
and save it to your configuration file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			stderr := cmd.ErrOrStderr()

			// 1. Print the token generation URL to stderr.
			_, _ = fmt.Fprintf(stderr, "Open the following URL to generate your API token:\n  %s\n\n", tokenURL)

			// 2. Try to open the browser — ignore errors (not all environments have one).
			openBrowser(tokenURL)

			// 3. Read the token from stdin.
			var token string
			if tty.IsTerminal() {
				// Interactive: mask input like a password prompt.
				_, _ = fmt.Fprint(stderr, "Paste your API token: ")
				raw, err := term.ReadPassword(int(os.Stdin.Fd()))
				_, _ = fmt.Fprintln(stderr) // newline after masked input
				if err != nil {
					return fmt.Errorf("reading token: %w", err)
				}
				token = strings.TrimSpace(string(raw))
			} else {
				// Non-TTY / piped: read a single line from stdin.
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					token = strings.TrimSpace(scanner.Text())
				}
				if err := scanner.Err(); err != nil {
					return fmt.Errorf("reading token from stdin: %w", err)
				}
			}

			if token == "" {
				return fmt.Errorf("no token provided")
			}

			// 4. Validate the token by pinging the accounts endpoint.
			tempClient, err := api.NewClient(
				"https://api.activtrak.com",
				token,
				Version,
				false,
				stderr,
				false,
				os.Stdout,
			)
			if err != nil {
				return fmt.Errorf("creating validation client: %w", err)
			}
			tempClient.HTTPClient.Timeout = 15 * time.Second

			if err := tempClient.AccountPing(cmd.Context()); err != nil {
				return fmt.Errorf("token validation failed: %w. Check your token and try again", err)
			}

			// 5. Save the profile with restricted permissions.
			cfg := &config.Config{
				Token:   token,
				BaseURL: "https://api.activtrak.com",
				Format:  "table",
				Timeout: 30 * time.Second,
			}
			if err := config.SaveProfile(profileFlag, cfg); err != nil {
				return fmt.Errorf("saving profile %q: %w", profileFlag, err)
			}

			// 6. Confirm success.
			_, _ = fmt.Fprintf(stderr, "Logged in to profile %q\n", profileFlag)
			return nil
		},
	}

	cmd.Flags().StringVar(&profileFlag, "profile", "default", "Profile name to save credentials to")
	return cmd
}

// openBrowser attempts to open url in the system default browser.
// Errors are intentionally ignored — not all environments have a browser.
func openBrowser(url string) {
	var browserCmd string
	switch runtime.GOOS {
	case "darwin":
		browserCmd = "open"
	case "windows":
		browserCmd = "start"
	default:
		browserCmd = "xdg-open"
	}
	_ = exec.Command(browserCmd, url).Start()
}
