package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/output"
)

// newNotificationsCmd returns the deprecated "notifications" resource command.
//
// Deprecated: Use 'atadmin signals list' instead.
func newNotificationsCmd(state *appState) *cobra.Command {
	notifications := &cobra.Command{
		Use:        "notifications",
		Short:      "Manage ActivTrak notifications (deprecated)",
		Deprecated: "Use 'atadmin signals list' instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	notifications.AddCommand(newNotificationsListCmd(state))
	return notifications
}

// newNotificationsListCmd implements "notifications list".
func newNotificationsListCmd(state *appState) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List notifications (deprecated — use 'signals list')",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Emit deprecation warning to stderr so it doesn't corrupt piped output.
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Warning: 'notifications list' is deprecated. Use 'atadmin signals list' instead.")

			signals, err := state.client.GetLegacyNotifications(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing notifications: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), signals)
			}

			rows := make([][]string, len(signals))
			for i, s := range signals {
				rows[i] = []string{
					strconv.Itoa(s.ID),
					s.Name,
					s.Type,
					strconv.FormatBool(s.Enabled),
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "NAME", "TYPE", "ENABLED"}, rows)
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	return cmd
}
