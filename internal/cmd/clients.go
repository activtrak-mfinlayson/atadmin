package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/api"
	"github.com/activtrak-mfinlayson/atadmin/internal/bulk"
	"github.com/activtrak-mfinlayson/atadmin/internal/output"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)


// newClientsCmd returns the "clients" resource command with all subcommands wired.
func newClientsCmd(state *appState) *cobra.Command {
	clients := &cobra.Command{
		Use:   "clients",
		Short: "Manage ActivTrak clients",
		Long:  `Commands for listing, inspecting, and managing tracked end-user clients.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	clients.AddCommand(newClientsListCmd(state))
	clients.AddCommand(newClientsGetCmd(state))
	clients.AddCommand(newClientsUpdateCmd(state))
	clients.AddCommand(newClientsHealthCmd(state))
	clients.AddCommand(newClientsDeleteCmd(state))
	clients.AddCommand(newClientsRestoreCmd(state))
	clients.AddCommand(newClientsMergeCmd(state))
	clients.AddCommand(newClientsMergeBulkCmd(state))
	clients.AddCommand(newClientsUnmergeBulkCmd(state))
	clients.AddCommand(newClientsAliasCmd(state))
	clients.AddCommand(newClientsDoNotTrackCmd(state))

	return clients
}

// newClientsListCmd implements "clients list".
func newClientsListCmd(state *appState) *cobra.Command {
	var (
		page        int
		pageSize    int
		asJSON      bool
		fieldsFlag  string
		summaryFlag bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all clients",
		RunE: func(cmd *cobra.Command, args []string) error {
			clients, err := state.client.ListClients(cmd.Context(), page, pageSize)
			if err != nil {
				return fmt.Errorf("listing clients: %w", err)
			}

			if asJSON {
				if summaryFlag {
					return output.JSONSummary(cmd.OutOrStdout(), len(clients), nil, len(clients) == pageSize)
				}
				if fieldsFlag != "" {
					generic, err := output.ToGeneric(clients)
					if err != nil {
						return fmt.Errorf("serializing results: %w", err)
					}
					return output.JSON(cmd.OutOrStdout(), output.FilterFields(generic, strings.Split(fieldsFlag, ",")))
				}
				return output.JSON(cmd.OutOrStdout(), clients)
			}

			rows := make([][]string, len(clients))
			for i, c := range clients {
				rows[i] = []string{strconv.Itoa(c.ID), c.Username, c.Alias, c.Status}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "USERNAME", "ALIAS", "STATUS"}, rows)
			return nil
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 50, "Number of results per page")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	cmd.Flags().StringVar(&fieldsFlag, "fields", "", "Comma-separated top-level JSON keys to include (e.g. id,username)")
	cmd.Flags().BoolVar(&summaryFlag, "summary", false, "Return aggregate statistics instead of full results")

	return cmd
}

// newClientsGetCmd implements "clients get <id-or-username>".
func newClientsGetCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-username>",
		Short: "Get a client by ID or username",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				client *api.ATClient
				err    error
			)

			if id, parseErr := strconv.Atoi(args[0]); parseErr == nil {
				client, err = state.client.GetClientByID(cmd.Context(), id)
			} else {
				client, err = state.client.GetClientByUsername(cmd.Context(), args[0])
			}
			if err != nil {
				return fmt.Errorf("getting client %q: %w", args[0], err)
			}

			output.KeyValue(cmd.OutOrStdout(), map[string]string{
				"id":          strconv.Itoa(client.ID),
				"username":    client.Username,
				"logonDomain": client.LogonDomain,
				"alias":       client.Alias,
				"status":      client.Status,
				"deviceCount": strconv.Itoa(client.DeviceCount),
			})
			return nil
		},
	}
}

// newClientsUpdateCmd implements "clients update <id> --alias <str>".
func newClientsUpdateCmd(state *appState) *cobra.Command {
	var alias string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a client's alias",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid client ID %q: must be an integer", args[0])
			}
			if alias == "" {
				return fmt.Errorf("--alias is required")
			}
			if err := state.client.UpdateClientAlias(cmd.Context(), id, alias); err != nil {
				return fmt.Errorf("updating client %d: %w", id, err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated client %d\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&alias, "alias", "", "New alias for the client (required)")
	return cmd
}

// newClientsHealthCmd implements "clients health".
func newClientsHealthCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Show count of active clients",
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := state.client.ClientHealth(cmd.Context())
			if err != nil {
				return fmt.Errorf("fetching client health: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Active clients: %d\n", n)
			return nil
		},
	}
}

// newClientsDeleteCmd implements "clients delete --ids <ids>".
func newClientsDeleteCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete clients by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.DeleteClients(cmd.Context(), ids); err != nil {
				return fmt.Errorf("deleting clients: %w", err)
			}

			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted %d clients\n", len(ids))
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "deleted")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated client IDs to delete (required)")

	return cmd
}

// newClientsRestoreCmd implements "clients restore --ids <ids>".
func newClientsRestoreCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore deleted clients by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.RestoreClients(cmd.Context(), ids); err != nil {
				return fmt.Errorf("restoring clients: %w", err)
			}

			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Restored %d clients\n", len(ids))
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "restored")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated client IDs to restore (required)")

	return cmd
}

// newClientsMergeCmd implements "clients merge --source <id> --target <id>".
func newClientsMergeCmd(state *appState) *cobra.Command {
	var (
		sourceID int
		targetID int
	)

	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge one client into another",
		RunE: func(cmd *cobra.Command, args []string) error {
			if sourceID == 0 {
				return fmt.Errorf("--source is required")
			}
			if targetID == 0 {
				return fmt.Errorf("--target is required")
			}

			if err := state.client.MergeUsers(cmd.Context(), sourceID, targetID); err != nil {
				return fmt.Errorf("merging clients %d -> %d: %w", sourceID, targetID, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Merged client %d into %d\n", sourceID, targetID)
			return nil
		},
	}

	cmd.Flags().IntVar(&sourceID, "source", 0, "Source client ID (to be merged, required)")
	cmd.Flags().IntVar(&targetID, "target", 0, "Target client ID (to merge into, required)")

	return cmd
}

// newClientsMergeBulkCmd implements "clients merge-bulk --file <path>".
func newClientsMergeBulkCmd(state *appState) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "merge-bulk",
		Short: "Merge clients in bulk from a JSON or CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			records, err := bulk.ParseFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}

			if err := state.client.MergeUsersBulk(cmd.Context(), records); err != nil {
				return fmt.Errorf("bulk merging clients: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Merged %d client pairs\n", len(records))
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON or CSV file (required)")

	return cmd
}

// newClientsUnmergeBulkCmd implements "clients unmerge-bulk --file <path>".
func newClientsUnmergeBulkCmd(state *appState) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "unmerge-bulk",
		Short: "Unmerge clients in bulk from a JSON or CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			records, err := bulk.ParseFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}

			if err := state.client.UnmergeUsersBulk(cmd.Context(), records); err != nil {
				return fmt.Errorf("bulk unmerging clients: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Unmerged %d client records\n", len(records))
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON or CSV file (required)")

	return cmd
}

// newClientsAliasCmd returns the "clients alias" subcommand group.
func newClientsAliasCmd(state *appState) *cobra.Command {
	alias := &cobra.Command{
		Use:   "alias",
		Short: "Manage client aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	alias.AddCommand(newClientsAliasSetCmd(state))
	alias.AddCommand(newClientsAliasBulkCmd(state))

	return alias
}

// newClientsAliasSetCmd implements "clients alias set --id <id> --alias <str>".
func newClientsAliasSetCmd(state *appState) *cobra.Command {
	var (
		clientID  int
		aliasName string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set the alias for a client",
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientID == 0 {
				return fmt.Errorf("--id is required")
			}
			if aliasName == "" {
				return fmt.Errorf("--alias is required")
			}

			if err := state.client.UpdateAlias(cmd.Context(), clientID, aliasName); err != nil {
				return fmt.Errorf("setting alias for client %d: %w", clientID, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated alias for client %d\n", clientID)
			return nil
		},
	}

	cmd.Flags().IntVar(&clientID, "id", 0, "Client ID (required)")
	cmd.Flags().StringVar(&aliasName, "alias", "", "Alias string to assign (required)")

	return cmd
}

// newClientsAliasBulkCmd implements "clients alias bulk --file <path>".
func newClientsAliasBulkCmd(state *appState) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Update client aliases in bulk from a JSON or CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			records, err := bulk.ParseFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}

			if err := state.client.UpdateAliasBulk(cmd.Context(), records); err != nil {
				return fmt.Errorf("bulk updating aliases: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated aliases for %d clients\n", len(records))
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON or CSV file (required)")

	return cmd
}

// newClientsDoNotTrackCmd returns the "clients donottrack" subcommand group.
func newClientsDoNotTrackCmd(state *appState) *cobra.Command {
	dnt := &cobra.Command{
		Use:   "donottrack",
		Short: "Manage Do Not Track rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	dnt.AddCommand(newDNTListCmd(state))
	dnt.AddCommand(newDNTAddCmd(state))
	dnt.AddCommand(newDNTRemoveCmd(state))
	dnt.AddCommand(newDNTUpdateCmd(state))
	dnt.AddCommand(newDNTAddBulkCmd(state))
	dnt.AddCommand(newDNTRemoveBulkCmd(state))
	dnt.AddCommand(newDNTGlobalUserCmd(state))

	return dnt
}

// newDNTListCmd implements "clients donottrack list".
func newDNTListCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all Do Not Track entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := state.client.ListDoNotTrack(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing Do Not Track entries: %w", err)
			}

			rows := make([][]string, len(entries))
			for i, e := range entries {
				rows[i] = []string{
					strconv.Itoa(e.ID),
					e.LogonDomain,
					e.Username,
					strconv.FormatBool(e.IsGlobal),
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "DOMAIN", "USERNAME", "GLOBAL"}, rows)
			return nil
		},
	}
}

// newDNTAddCmd implements "clients donottrack add --domain <str> --username <str>".
func newDNTAddCmd(state *appState) *cobra.Command {
	var (
		domain   string
		username string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a Do Not Track entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			if domain == "" && username == "" {
				return fmt.Errorf("at least one of --domain or --username is required")
			}

			if err := state.client.AddDoNotTrack(cmd.Context(), domain, username); err != nil {
				return fmt.Errorf("adding Do Not Track entry: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Do Not Track entry added")
			return nil
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Logon domain to exclude")
	cmd.Flags().StringVar(&username, "username", "", "Username to exclude")

	return cmd
}

// newDNTRemoveCmd implements "clients donottrack remove --ids <ids>".
func newDNTRemoveCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove Do Not Track entries by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.RemoveDoNotTrack(cmd.Context(), ids); err != nil {
				return fmt.Errorf("removing Do Not Track entries: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed %d Do Not Track entries\n", len(ids))
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated Do Not Track entry IDs to remove (required)")

	return cmd
}

// newDNTUpdateCmd implements "clients donottrack update --id <id> --domain <str> --username <str>".
func newDNTUpdateCmd(state *appState) *cobra.Command {
	var (
		entryID  int
		domain   string
		username string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a Do Not Track entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			if entryID == 0 {
				return fmt.Errorf("--id is required")
			}

			if err := state.client.UpdateDoNotTrack(cmd.Context(), entryID, domain, username); err != nil {
				return fmt.Errorf("updating Do Not Track entry %d: %w", entryID, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated Do Not Track entry %d\n", entryID)
			return nil
		},
	}

	cmd.Flags().IntVar(&entryID, "id", 0, "Do Not Track entry ID (required)")
	cmd.Flags().StringVar(&domain, "domain", "", "New logon domain value")
	cmd.Flags().StringVar(&username, "username", "", "New username value")

	return cmd
}

// newDNTAddBulkCmd implements "clients donottrack add-bulk --file <path>".
func newDNTAddBulkCmd(state *appState) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "add-bulk",
		Short: "Add Do Not Track entries in bulk from a JSON or CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			records, err := bulk.ParseFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}

			if err := state.client.AddDoNotTrackBulk(cmd.Context(), records); err != nil {
				return fmt.Errorf("bulk adding Do Not Track entries: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added %d Do Not Track entries\n", len(records))
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON or CSV file (required)")

	return cmd
}

// newDNTRemoveBulkCmd implements "clients donottrack remove-bulk --file <path>".
func newDNTRemoveBulkCmd(state *appState) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "remove-bulk",
		Short: "Remove Do Not Track entries in bulk from a JSON or CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			records, err := bulk.ParseFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}

			if err := state.client.RemoveDoNotTrackBulk(cmd.Context(), records); err != nil {
				return fmt.Errorf("bulk removing Do Not Track entries: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed %d Do Not Track entries\n", len(records))
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON or CSV file (required)")

	return cmd
}

// newDNTGlobalUserCmd implements "clients donottrack global-user --ids <ids>".
func newDNTGlobalUserCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "global-user",
		Short: "Mark Do Not Track entries as global by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.MarkGlobalUser(cmd.Context(), ids); err != nil {
				return fmt.Errorf("marking global Do Not Track entries: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Marked %d entries as global\n", len(ids))
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated Do Not Track entry IDs to mark as global (required)")

	return cmd
}

// parseIDs splits a comma-separated string of integers into []int.
// Empty string returns an empty (not nil) slice. Non-integer tokens return an error.
func parseIDs(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []int{}, nil
	}

	parts := strings.Split(s, ",")
	ids := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid ID %q: must be an integer", p)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
