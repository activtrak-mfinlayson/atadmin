// Package cmd contains all Cobra command definitions for atadmin.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/api"
	"github.com/activtrak-mfinlayson/atadmin/internal/config"
)

// Version is the current atadmin release version.
const Version = "0.1.0"

// appState carries resolved config and API client down to subcommands.
type appState struct {
	cfg    *config.Config
	client *api.Client
}

// NewRootCmd constructs and returns the atadmin root command.
// A fresh command is created on each call, enabling safe use in tests.
func NewRootCmd() *cobra.Command {
	var (
		profileFlag string
		formatFlag  string
		tokenFlag   string
		baseURLFlag string
		verboseFlag bool
		versionFlag bool
	)

	var state appState

	root := &cobra.Command{
		Use:           "atadmin",
		Short:         "ActivTrak admin CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
		Long: `atadmin is the ActivTrak administration command-line interface.

Use it to manage users, groups, devices, and settings for your ActivTrak account.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadProfile(profileFlag)
			if err != nil {
				return fmt.Errorf("loading profile %q: %w", profileFlag, err)
			}
			// Flag overrides beat config file.
			if tokenFlag != "" {
				cfg.Token = tokenFlag
			}
			if baseURLFlag != "" {
				cfg.BaseURL = baseURLFlag
			}
			if formatFlag != "" {
				cfg.Format = formatFlag
			}

			client, err := api.NewClient(cfg.BaseURL, cfg.Token, Version, verboseFlag, os.Stderr)
			if err != nil {
				return fmt.Errorf("creating API client: %w", err)
			}
			client.HTTPClient.Timeout = cfg.Timeout

			state.cfg = cfg
			state.client = client
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "atadmin version %s\n", Version)
				return nil
			}
			return cmd.Help()
		},
	}

	root.Flags().BoolVar(&versionFlag, "version", false, "Show version information")
	root.PersistentFlags().StringVar(&profileFlag, "profile", "default", "Config profile to use (env: ATADMIN_PROFILE)")
	root.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "Print HTTP request/response details to stderr")
	root.PersistentFlags().StringVarP(&formatFlag, "format", "f", "", "Output format: table or json (env: ATADMIN_FORMAT)")
	root.PersistentFlags().StringVar(&tokenFlag, "token", "", "API bearer token (env: ATADMIN_TOKEN)")
	root.PersistentFlags().StringVar(&baseURLFlag, "base-url", "", "API base URL (env: ATADMIN_BASE_URL)")

	// Register all resource subcommands.
	root.AddCommand(newAuthCmd())
	root.AddCommand(newUsersCmd(&state))
	root.AddCommand(newAgentsCmd(&state))
	root.AddCommand(newClientsCmd(&state))
	root.AddCommand(newGroupsCmd(&state))
	root.AddCommand(newConsumersCmd(&state))
	root.AddCommand(newDevicesCmd(&state))
	root.AddCommand(newSettingsCmd(&state))
	root.AddCommand(newAlarmsCmd(&state))
	root.AddCommand(newSignalsCmd(&state))
	root.AddCommand(newSchedulesCmd(&state))
	root.AddCommand(newAPIKeysCmd(&state))
	root.AddCommand(newAuditLogCmd(&state))
	root.AddCommand(newHRDCCmd(&state))
	root.AddCommand(newNotificationsCmd(&state))
	root.AddCommand(newMCPCmd())

	return root
}

// Execute runs atadmin and writes errors to the command's error output.
func Execute() error {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintf(root.ErrOrStderr(), "Error: %s\nRun 'atadmin --help' for usage.\n", err)
		return err
	}
	return nil
}
