package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/output"
)

// newAuditLogCmd returns the "auditlog" resource command with all subcommands wired.
func newAuditLogCmd(state *appState) *cobra.Command {
	auditlog := &cobra.Command{
		Use:   "auditlog",
		Short: "Query the ActivTrak audit log",
		Long:  `Commands for listing audit log entries and retrieving attachments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	auditlog.AddCommand(newAuditLogListCmd(state))
	auditlog.AddCommand(newAuditLogAttachmentCmd(state))

	return auditlog
}

// newAuditLogListCmd implements "auditlog list".
func newAuditLogListCmd(state *appState) *cobra.Command {
	var (
		from     string
		to       string
		filters  string
		sortCol  string
		sortDesc bool
		page     int
		pageSize int
		asJSON   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List audit log entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			logs, err := state.client.ListAuditLogs(
				cmd.Context(),
				from, to, filters, sortCol,
				sortDesc,
				page, pageSize,
			)
			if err != nil {
				return fmt.Errorf("listing audit logs: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), logs)
			}

			rows := make([][]string, len(logs))
			for i, l := range logs {
				rows[i] = []string{
					strconv.Itoa(l.ID),
					l.Action,
					l.Actor,
					l.Timestamp,
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "ACTION", "ACTOR", "TIMESTAMP"}, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Filter by start date/time (e.g. 2024-01-01)")
	cmd.Flags().StringVar(&to, "to", "", "Filter by end date/time (e.g. 2024-01-31)")
	cmd.Flags().StringVar(&filters, "filters", "", "Additional filter expression")
	cmd.Flags().StringVar(&sortCol, "sort", "", "Column to sort by")
	cmd.Flags().BoolVar(&sortDesc, "desc", false, "Sort in descending order")
	cmd.Flags().IntVar(&page, "page", 0, "Page number (0 = server default)")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "Results per page (0 = server default)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")

	return cmd
}

// newAuditLogAttachmentCmd returns the "auditlog attachment" subcommand group.
func newAuditLogAttachmentCmd(state *appState) *cobra.Command {
	attachment := &cobra.Command{
		Use:   "attachment",
		Short: "Manage audit log attachments",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	attachment.AddCommand(newAuditLogAttachmentGetCmd(state))
	return attachment
}

// newAuditLogAttachmentGetCmd implements "auditlog attachment get <id> [--output <path>]".
func newAuditLogAttachmentGetCmd(state *appState) *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Download an audit log attachment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			data, err := state.client.GetAttachment(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("getting attachment %q: %w", id, err)
			}

			if outputPath == "" {
				// Write raw bytes to stdout.
				_, err = cmd.OutOrStdout().Write(data)
				return err
			}

			if err := os.WriteFile(outputPath, data, 0600); err != nil {
				return fmt.Errorf("writing attachment to %q: %w", outputPath, err)
			}
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Attachment saved to %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Path to save the attachment (default: stdout)")
	return cmd
}
