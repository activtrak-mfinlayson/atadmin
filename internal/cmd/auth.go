package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the ActivTrak API",
		Long: `Open your browser to the API token generation page, paste your token,
and save it to your configuration file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Not yet implemented. Run 'atadmin auth --help' for usage.")
			return fmt.Errorf("not implemented")
		},
	}
}
