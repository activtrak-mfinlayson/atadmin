// export_test.go provides test-only helpers in the cmd package so that
// the external cmd_test package can construct lightweight command trees
// without triggering the real PersistentPreRunE config/API setup.
package cmd

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/api"
	"github.com/activtrak-mfinlayson/atadmin/internal/config"
)

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

// NewTestRootWithDryRun returns a minimal root command with a pre-configured
// api.Client (DryRun=true, Out=dryRunOut) and only the "groups" subtree
// attached. No PersistentPreRunE — tests never need a config file or token.
func NewTestRootWithDryRun(dryRunOut io.Writer) *cobra.Command {
	client, _ := api.NewClient("http://fake.invalid", "tok", "0.1.0", false, nil, true, dryRunOut)
	state := &appState{
		client: client,
		cfg:    &config.Config{Format: "table"},
	}
	root := &cobra.Command{
		Use:           "atadmin",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(newGroupsCmd(state))
	return root
}
