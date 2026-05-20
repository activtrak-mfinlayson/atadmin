package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/activtrak-mfinlayson/atadmin/internal/bulk"
	"github.com/activtrak-mfinlayson/atadmin/internal/output"
	"github.com/activtrak-mfinlayson/atadmin/internal/stdin"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)

// newConsumersCmd returns the "consumers" resource command with all subcommands wired.
func newConsumersCmd(state *appState) *cobra.Command {
	consumers := &cobra.Command{
		Use:   "consumers",
		Short: "Manage ActivTrak consumers",
		Long:  `Commands for listing, inspecting, and managing account consumers (report/admin users).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	consumers.AddCommand(newConsumersListCmd(state))
	consumers.AddCommand(newConsumersGetCmd(state))
	consumers.AddCommand(newConsumersCreateCmd(state))
	consumers.AddCommand(newConsumersUpdateCmd(state))
	consumers.AddCommand(newConsumersDeleteCmd(state))
	consumers.AddCommand(newConsumersDeleteBulkCmd(state))
	consumers.AddCommand(newConsumersRoleCmd(state))
	consumers.AddCommand(newConsumersPasswordCmd(state))
	consumers.AddCommand(newConsumersSSOCmd(state))
	consumers.AddCommand(newConsumersGroupsCmd(state))
	consumers.AddCommand(newConsumersChromeUsersCmd(state))

	return consumers
}

// newConsumersListCmd implements "consumers list".
func newConsumersListCmd(state *appState) *cobra.Command {
	var (
		page        int
		pageSize    int
		asJSON      bool
		fieldsFlag  string
		summaryFlag bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all consumers",
		RunE: func(cmd *cobra.Command, args []string) error {
			consumers, err := state.client.ListConsumers(cmd.Context(), page, pageSize)
			if err != nil {
				return fmt.Errorf("listing consumers: %w", err)
			}

			if asJSON {
				if summaryFlag {
					return output.JSONSummary(cmd.OutOrStdout(), len(consumers), nil, len(consumers) == pageSize)
				}
				if fieldsFlag != "" {
					generic, err := output.ToGeneric(consumers)
					if err != nil {
						return fmt.Errorf("serializing results: %w", err)
					}
					return output.JSON(cmd.OutOrStdout(), output.FilterFields(generic, strings.Split(fieldsFlag, ",")))
				}
				return output.JSON(cmd.OutOrStdout(), consumers)
			}

			rows := make([][]string, len(consumers))
			for i, c := range consumers {
				rows[i] = []string{
					strconv.Itoa(c.ID),
					c.Username,
					c.Role,
					strconv.FormatBool(c.UseSSO),
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "USERNAME", "ROLE", "SSO"}, rows)
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

// newConsumersGetCmd implements "consumers get <id>".
func newConsumersGetCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a consumer by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid consumer ID %q: must be an integer", args[0])
			}

			consumer, err := state.client.GetConsumer(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("getting consumer %d: %w", id, err)
			}

			output.KeyValue(cmd.OutOrStdout(), map[string]string{
				"id":       strconv.Itoa(consumer.ID),
				"username": consumer.Username,
				"role":     consumer.Role,
				"useSSO":   strconv.FormatBool(consumer.UseSSO),
			})
			return nil
		},
	}
}

// newConsumersCreateCmd implements "consumers create --file <path>".
func newConsumersCreateCmd(state *appState) *cobra.Command {
	var (
		filePath  string
		fromStdin bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create consumers from a JSON or CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath != "" && fromStdin {
				return fmt.Errorf("--from-stdin: --file and --from-stdin are mutually exclusive")
			}

			var records []map[string]any
			var err error

			if fromStdin {
				if err := stdin.EnsurePiped(); err != nil {
					return err
				}
				records, err = stdin.ReadRecords(os.Stdin)
				if err != nil {
					return err
				}
			} else {
				if filePath == "" {
					return fmt.Errorf("--file is required")
				}
				records, err = bulk.ParseFile(filePath)
				if err != nil {
					return fmt.Errorf("reading file %q: %w", filePath, err)
				}
			}

			if err := state.client.CreateConsumers(cmd.Context(), records); err != nil {
				return fmt.Errorf("creating consumers: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Created %d consumers\n", len(records))
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON or CSV file")
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read JSON array from stdin instead of --file")

	return cmd
}

// newConsumersUpdateCmd implements "consumers update --file <path>".
func newConsumersUpdateCmd(state *appState) *cobra.Command {
	var (
		filePath  string
		fromStdin bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update consumers from a JSON or CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath != "" && fromStdin {
				return fmt.Errorf("--from-stdin: --file and --from-stdin are mutually exclusive")
			}

			var records []map[string]any
			var err error

			if fromStdin {
				if err := stdin.EnsurePiped(); err != nil {
					return err
				}
				records, err = stdin.ReadRecords(os.Stdin)
				if err != nil {
					return err
				}
			} else {
				if filePath == "" {
					return fmt.Errorf("--file is required")
				}
				records, err = bulk.ParseFile(filePath)
				if err != nil {
					return fmt.Errorf("reading file %q: %w", filePath, err)
				}
			}

			if err := state.client.PatchConsumers(cmd.Context(), records); err != nil {
				return fmt.Errorf("updating consumers: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Updated %d consumers\n", len(records))
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON or CSV file")
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read JSON array from stdin instead of --file")

	return cmd
}

// newConsumersDeleteCmd implements "consumers delete --ids <csv>".
func newConsumersDeleteCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete consumers by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.DeleteConsumers(cmd.Context(), ids); err != nil {
				return fmt.Errorf("deleting consumers: %w", err)
			}

			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Deleted %d consumers\n", len(ids))
			} else {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "deleted")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated consumer IDs to delete (required)")

	return cmd
}

// newConsumersDeleteBulkCmd implements "consumers delete-bulk --file <path>".
func newConsumersDeleteBulkCmd(state *appState) *cobra.Command {
	var (
		filePath  string
		fromStdin bool
	)

	cmd := &cobra.Command{
		Use:   "delete-bulk",
		Short: "Delete consumers in bulk from a JSON or CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath != "" && fromStdin {
				return fmt.Errorf("--from-stdin: --file and --from-stdin are mutually exclusive")
			}

			var records []map[string]any
			var err error

			if fromStdin {
				if err := stdin.EnsurePiped(); err != nil {
					return err
				}
				records, err = stdin.ReadRecords(os.Stdin)
				if err != nil {
					return err
				}
			} else {
				if filePath == "" {
					return fmt.Errorf("--file is required")
				}
				records, err = bulk.ParseFile(filePath)
				if err != nil {
					return fmt.Errorf("reading file %q: %w", filePath, err)
				}
			}

			if err := state.client.DeleteConsumersBulk(cmd.Context(), records); err != nil {
				return fmt.Errorf("bulk deleting consumers: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Deleted %d consumers\n", len(records))
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON or CSV file")
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read JSON array from stdin instead of --file")

	return cmd
}

// newConsumersRoleCmd returns the "consumers role" subcommand group.
func newConsumersRoleCmd(state *appState) *cobra.Command {
	role := &cobra.Command{
		Use:   "role",
		Short: "Manage consumer roles",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	role.AddCommand(newConsumersRoleSetCmd(state))

	return role
}

// newConsumersRoleSetCmd implements "consumers role set <id> --role <str>".
func newConsumersRoleSetCmd(state *appState) *cobra.Command {
	var role string

	cmd := &cobra.Command{
		Use:   "set <id>",
		Short: "Set the role for a consumer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid consumer ID %q: must be an integer", args[0])
			}
			if role == "" {
				return fmt.Errorf("--role is required")
			}

			if err := state.client.SetConsumerRole(cmd.Context(), id, role); err != nil {
				return fmt.Errorf("setting role for consumer %d: %w", id, err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Set role for consumer %d to %q\n", id, role)
			return nil
		},
	}

	cmd.Flags().StringVar(&role, "role", "", "Role to assign (required)")

	return cmd
}

// newConsumersPasswordCmd returns the "consumers password" subcommand group.
func newConsumersPasswordCmd(state *appState) *cobra.Command {
	pw := &cobra.Command{
		Use:   "password",
		Short: "Manage consumer passwords",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	pw.AddCommand(newConsumersPasswordSetCmd(state))

	return pw
}

type passwordStdinPayload struct {
	Password string `json:"password"`
}

// newConsumersPasswordSetCmd implements "consumers password set <id>".
// Prompts for masked input in interactive mode; rejects non-TTY invocations unless --from-stdin is set.
func newConsumersPasswordSetCmd(state *appState) *cobra.Command {
	var fromStdin bool

	cmd := &cobra.Command{
		Use:   "set <id>",
		Short: "Set the password for a consumer (interactive prompt)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid consumer ID %q: must be an integer", args[0])
			}

			var password string

			if fromStdin {
				if err := stdin.EnsurePiped(); err != nil {
					return err
				}
				payload, err := stdin.ReadJSON[passwordStdinPayload](os.Stdin)
				if err != nil {
					return err
				}
				if payload.Password == "" {
					return fmt.Errorf("--from-stdin: \"password\" field is required")
				}
				password = payload.Password
			} else {
				if !tty.IsTerminal() {
					return fmt.Errorf("password prompt requires interactive terminal")
				}

				_, _ = fmt.Fprint(cmd.ErrOrStderr(), "New password: ")
				password, err = readPassword(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading password: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.ErrOrStderr())

				if password == "" {
					return fmt.Errorf("password must not be empty")
				}
			}

			if err := state.client.SetConsumerPassword(cmd.Context(), id, password); err != nil {
				return fmt.Errorf("setting password for consumer %d: %w", id, err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Password updated for consumer %d\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read password as JSON {\"password\":\"...\"} from stdin")
	return cmd
}

// readPassword reads a masked password from f. When f is a real terminal it
// uses term.ReadPassword for echo suppression; otherwise it falls back to
// line-buffered reading (useful in tests).
func readPassword(f *os.File) (string, error) {
	if term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	// Fallback for non-terminal file descriptors (e.g. pipe in tests).
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

// newConsumersSSOCmd returns the "consumers sso" subcommand group.
func newConsumersSSOCmd(state *appState) *cobra.Command {
	sso := &cobra.Command{
		Use:   "sso",
		Short: "Manage consumer SSO settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	sso.AddCommand(newConsumersSSOSetCmd(state))

	return sso
}

// newConsumersSSOSetCmd implements "consumers sso set --consumer <id> --use-sso <bool>".
func newConsumersSSOSetCmd(state *appState) *cobra.Command {
	var (
		consumerID int
		useSSO     bool
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Enable or disable SSO for a consumer",
		RunE: func(cmd *cobra.Command, args []string) error {
			if consumerID == 0 {
				return fmt.Errorf("--consumer is required")
			}

			if err := state.client.SetConsumerSSO(cmd.Context(), consumerID, useSSO); err != nil {
				return fmt.Errorf("setting SSO for consumer %d: %w", consumerID, err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Set SSO=%v for consumer %d\n", useSSO, consumerID)
			return nil
		},
	}

	cmd.Flags().IntVar(&consumerID, "consumer", 0, "Consumer ID (required)")
	cmd.Flags().BoolVar(&useSSO, "use-sso", false, "Enable (true) or disable (false) SSO")

	return cmd
}

// newConsumersGroupsCmd returns the "consumers groups" subcommand group.
func newConsumersGroupsCmd(state *appState) *cobra.Command {
	groups := &cobra.Command{
		Use:   "groups",
		Short: "Manage viewable groups for a consumer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	groups.AddCommand(newConsumersGroupsAddCmd(state))

	return groups
}

// newConsumersGroupsAddCmd implements "consumers groups add <id> --group-ids <csv>".
func newConsumersGroupsAddCmd(state *appState) *cobra.Command {
	var groupIDsFlag string

	cmd := &cobra.Command{
		Use:   "add <id>",
		Short: "Add viewable groups for a consumer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid consumer ID %q: must be an integer", args[0])
			}

			groupIDs, err := parseIDs(groupIDsFlag)
			if err != nil {
				return fmt.Errorf("parsing --group-ids: %w", err)
			}
			if len(groupIDs) == 0 {
				return fmt.Errorf("--group-ids is required")
			}

			if err := state.client.AddConsumerViewableGroups(cmd.Context(), id, groupIDs); err != nil {
				return fmt.Errorf("adding viewable groups for consumer %d: %w", id, err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Added %d viewable group(s) for consumer %d\n", len(groupIDs), id)
			return nil
		},
	}

	cmd.Flags().StringVar(&groupIDsFlag, "group-ids", "", "Comma-separated group IDs to add (required)")

	return cmd
}

// newConsumersChromeUsersCmd returns the "consumers chrome-users" subcommand group.
func newConsumersChromeUsersCmd(state *appState) *cobra.Command {
	cu := &cobra.Command{
		Use:   "chrome-users",
		Short: "Manage Chrome-managed users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cu.AddCommand(newConsumersChromeUsersImportCmd(state))

	return cu
}

// newConsumersChromeUsersImportCmd implements "consumers chrome-users import --file <path>".
func newConsumersChromeUsersImportCmd(state *appState) *cobra.Command {
	var (
		filePath  string
		fromStdin bool
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import Chrome users in bulk from a JSON or CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath != "" && fromStdin {
				return fmt.Errorf("--from-stdin: --file and --from-stdin are mutually exclusive")
			}

			var records []map[string]any
			var err error

			if fromStdin {
				if err := stdin.EnsurePiped(); err != nil {
					return err
				}
				records, err = stdin.ReadRecords(os.Stdin)
				if err != nil {
					return err
				}
			} else {
				if filePath == "" {
					return fmt.Errorf("--file is required")
				}
				records, err = bulk.ParseFile(filePath)
				if err != nil {
					return fmt.Errorf("reading file %q: %w", filePath, err)
				}
			}

			if err := state.client.CreateChromeUsersBulk(cmd.Context(), records); err != nil {
				return fmt.Errorf("importing Chrome users: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Imported %d Chrome user records\n", len(records))
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON or CSV file")
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read JSON array from stdin instead of --file")

	return cmd
}
