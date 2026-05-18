// Package cmd contains all Cobra command definitions for atadmin.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current atadmin release version.
const Version = "0.1.0"

// NewRootCmd constructs and returns the atadmin root command.
// A fresh command is created on each call, enabling safe use in tests.
func NewRootCmd() *cobra.Command {
	var (
		formatFlag  string
		tokenFlag   string
		baseURLFlag string
		versionFlag bool
	)

	root := &cobra.Command{
		Use:           "atadmin",
		Short:         "ActivTrak admin CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
		Long: `atadmin is the ActivTrak administration command-line interface.

Use it to manage users, projects, and settings for your ActivTrak account.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "atadmin version %s\n", Version)
				return nil
			}
			return cmd.Help()
		},
	}

	root.Flags().BoolVar(&versionFlag, "version", false, "Show version information")
	root.PersistentFlags().StringVarP(&formatFlag, "format", "f", "table", "Output format: table or json (env: ATADMIN_FORMAT)")
	root.PersistentFlags().StringVar(&tokenFlag, "token", "", "API bearer token (env: ATADMIN_TOKEN)")
	root.PersistentFlags().StringVar(&baseURLFlag, "base-url", "", "API base URL (env: ATADMIN_BASE_URL)")

	root.AddCommand(newAuthCmd())

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
