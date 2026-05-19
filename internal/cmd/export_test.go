// export_test.go provides test-only helpers in the cmd package so that
// the external cmd_test package can construct lightweight command trees
// without triggering the real PersistentPreRunE config/API setup.
package cmd

import "github.com/spf13/cobra"

// NewTestClientsRoot returns a minimal *cobra.Command that has only the
// "clients" subtree attached, with no PersistentPreRunE so tests never
// need a config file or API token.
func NewTestClientsRoot() *cobra.Command {
	state := &appState{} // zero-value; API methods not called in help/flag tests

	root := &cobra.Command{
		Use:           "atadmin",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	root.AddCommand(newClientsCmd(state))
	return root
}
