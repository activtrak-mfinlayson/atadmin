package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/api"
	"github.com/activtrak-mfinlayson/atadmin/internal/output"
	"github.com/activtrak-mfinlayson/atadmin/internal/stdin"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)

// newUsersCmd returns the "users" resource command with all subcommands wired.
func newUsersCmd(state *appState) *cobra.Command {
	users := &cobra.Command{
		Use:   "users",
		Short: "Manage ActivTrak identity users",
		Long:  `Commands for listing, inspecting, and managing identity entities (users).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	users.AddCommand(newUsersListCmd(state))
	users.AddCommand(newUsersGetCmd(state))
	users.AddCommand(newUsersUpdateCmd(state))
	users.AddCommand(newUsersDeleteCmd(state))
	users.AddCommand(newUsersGroupsCmd(state))
	users.AddCommand(newUsersBulkCmd(state))

	return users
}

// ---------------------------------------------------------------------------
// users list
// ---------------------------------------------------------------------------

func newUsersListCmd(state *appState) *cobra.Command {
	var (
		filter      string
		search      string
		searchType  string
		sort        string
		sortDir     string
		limit       int
		cursor      string
		asJSON      bool
		fieldsFlag  string
		summaryFlag bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List identity users",
		RunE: func(cmd *cobra.Command, args []string) error {
			if asJSON && !cmd.Flags().Changed("limit") && limit == 0 {
				limit = 50
			}

			params := api.IdentityListParams{
				Search:     search,
				SearchType: searchType,
				Sort:       sort,
				SortDir:    sortDir,
				Limit:      limit,
				Cursor:     cursor,
			}
			if filter != "" {
				params.Filters = []string{filter}
			}

			page, err := state.client.ListUsers(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("listing users: %w", err)
			}

			if asJSON {
				if summaryFlag {
					return output.JSONSummary(cmd.OutOrStdout(), len(page.Results), nil, page.Cursor != "")
				}
				if fieldsFlag != "" {
					generic, err := output.ToGeneric(page.Results)
					if err != nil {
						return fmt.Errorf("serializing results: %w", err)
					}
					return output.JSON(cmd.OutOrStdout(), output.FilterFields(generic, strings.Split(fieldsFlag, ",")))
				}
				return output.JSON(cmd.OutOrStdout(), page)
			}

			rows := make([][]string, len(page.Results))
			for i, u := range page.Results {
				groupNames := identityGroupNames(u.Groups)
				rows[i] = []string{
					strconv.FormatInt(u.ID, 10),
					api.FieldValue(u.DisplayName),
					u.Status,
					groupNames,
					strconv.FormatBool(u.Tracked),
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "DISPLAY NAME", "STATUS", "GROUPS", "TRACKED"}, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter type (tracked, untracked, active30days, ...)")
	cmd.Flags().StringVar(&search, "search", "", "Search term")
	cmd.Flags().StringVar(&searchType, "search-type", "", "Field to search (email, upn, displayname, alias, ...)")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort field")
	cmd.Flags().StringVar(&sortDir, "sort-dir", "", "Sort direction: asc or desc")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum results (0 = server default)")
	cmd.Flags().StringVar(&cursor, "cursor", "", "Pagination cursor from a previous response")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	cmd.Flags().StringVar(&fieldsFlag, "fields", "", "Comma-separated top-level JSON keys to include (e.g. id,email)")
	cmd.Flags().BoolVar(&summaryFlag, "summary", false, "Return aggregate statistics instead of full results")

	return cmd
}

// ---------------------------------------------------------------------------
// users get
// ---------------------------------------------------------------------------

func newUsersGetCmd(state *appState) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get details for a single identity user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}

			u, err := state.client.GetUser(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("getting user: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), u)
			}

			output.KeyValue(cmd.OutOrStdout(), map[string]string{
				"id":          strconv.FormatInt(u.ID, 10),
				"revision":    strconv.FormatInt(u.Revision, 10),
				"displayName": api.FieldValue(u.DisplayName),
				"firstName":   api.FieldValue(u.FirstName),
				"lastName":    api.FieldValue(u.LastName),
				"emails":      identityFieldList(u.Emails),
				"upns":        identityFieldList(u.UPNs),
				"employeeIds": identityFieldList(u.EmployeeIDs),
				"groups":      identityGroupNames(u.Groups),
				"primaryGroup": u.PrimaryGroupName,
				"tracked":     strconv.FormatBool(u.Tracked),
				"status":      u.Status,
				"timezone":    api.FieldValue(u.Timezone),
				"agents":      identityAgentSummary(u.Agents),
				"created":     u.Created,
				"updated":     u.Updated,
			})
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	return cmd
}

// ---------------------------------------------------------------------------
// users update
// ---------------------------------------------------------------------------

func newUsersUpdateCmd(state *appState) *cobra.Command {
	var (
		displayName string
		firstName   string
		lastName    string
		timezone    string
		tracked     string // "true" or "false" — using string to detect when flag is set
		revision    int64
		asJSON      bool
		fromStdin   bool
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update fields on an identity user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}

			var req api.UpdateUserRequest

			if fromStdin {
				if err := stdin.EnsurePiped(); err != nil {
					return err
				}
				req, err = stdin.ReadJSON[api.UpdateUserRequest](os.Stdin)
				if err != nil {
					return err
				}
			} else {
				flagCount := 0
				if cmd.Flags().Changed("display-name") {
					req.DisplayName = &displayName
					flagCount++
				}
				if cmd.Flags().Changed("first-name") {
					req.FirstName = &firstName
					flagCount++
				}
				if cmd.Flags().Changed("last-name") {
					req.LastName = &lastName
					flagCount++
				}
				if cmd.Flags().Changed("timezone") {
					req.Timezone = &timezone
					flagCount++
				}
				if cmd.Flags().Changed("tracked") {
					t := tracked == "true"
					req.Tracked = &t
					flagCount++
				}
				if flagCount == 0 {
					return fmt.Errorf("at least one of --display-name, --first-name, --last-name, --timezone, --tracked must be provided")
				}
			}

			if revision == 0 {
				current, err := state.client.GetUser(cmd.Context(), id)
				if err != nil {
					return fmt.Errorf("fetching current revision: %w", err)
				}
				revision = current.Revision
			}

			updated, err := state.client.UpdateUser(cmd.Context(), id, revision, req)
			if err != nil {
				return fmt.Errorf("updating user: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), updated)
			}
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Updated user %d\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&displayName, "display-name", "", "New display name")
	cmd.Flags().StringVar(&firstName, "first-name", "", "New first name")
	cmd.Flags().StringVar(&lastName, "last-name", "", "New last name")
	cmd.Flags().StringVar(&timezone, "timezone", "", "New timezone (IANA, e.g. America/Chicago)")
	cmd.Flags().StringVar(&tracked, "tracked", "", "Set tracking state: true or false")
	cmd.Flags().Int64Var(&revision, "revision", 0, "Explicit revision (skips auto-fetch)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output updated entity as JSON")
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read JSON payload from stdin (replaces individual field flags)")

	return cmd
}

// ---------------------------------------------------------------------------
// users delete
// ---------------------------------------------------------------------------

func newUsersDeleteCmd(state *appState) *cobra.Command {
	var (
		revision  int64
		yes       bool
		fromStdin bool
	)

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an identity user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}

			skipConfirmation := yes || fromStdin

			if !skipConfirmation {
				if !tty.IsTerminal() {
					return fmt.Errorf("non-interactive mode: pass --yes to confirm deletion")
				}
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Delete user %d? [y/N] ", id)
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if answer != "y" && answer != "yes" {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
					return nil
				}
			}

			if revision == 0 {
				current, err := state.client.GetUser(cmd.Context(), id)
				if err != nil {
					return fmt.Errorf("fetching current revision: %w", err)
				}
				revision = current.Revision
			}

			if err := state.client.DeleteUser(cmd.Context(), id, revision); err != nil {
				return fmt.Errorf("deleting user: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Deleted user %d\n", id)
			return nil
		},
	}

	cmd.Flags().Int64Var(&revision, "revision", 0, "Explicit revision (skips auto-fetch)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Skip confirmation prompt (no data read from stdin)")

	return cmd
}

// ---------------------------------------------------------------------------
// users groups
// ---------------------------------------------------------------------------

func newUsersGroupsCmd(state *appState) *cobra.Command {
	groups := &cobra.Command{
		Use:   "groups",
		Short: "Manage group memberships for an identity user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	groups.AddCommand(newUsersGroupsAddCmd(state))
	groups.AddCommand(newUsersGroupsRemoveCmd(state))
	return groups
}

func newUsersGroupsAddCmd(state *appState) *cobra.Command {
	var (
		groupIDsFlag string
		revision     int64
	)

	cmd := &cobra.Command{
		Use:   "add <userId> [groupId]",
		Short: "Add groups to an identity user",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid userId %q: %w", args[0], err)
			}

			var groupIDs []int
			switch {
			case len(args) == 2:
				n, err := strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid groupId %q: %w", args[1], err)
				}
				groupIDs = []int{n}
			case groupIDsFlag != "":
				var parseErr error
				groupIDs, parseErr = api.ParseGroupIDs(groupIDsFlag)
				if parseErr != nil {
					return parseErr
				}
			default:
				return fmt.Errorf("provide a groupId argument or --group-ids flag")
			}

			if revision == 0 {
				current, err := state.client.GetUser(cmd.Context(), userID)
				if err != nil {
					return fmt.Errorf("fetching current revision: %w", err)
				}
				revision = current.Revision
			}

			if _, err := state.client.AddUserGroups(cmd.Context(), userID, groupIDs, revision); err != nil {
				return fmt.Errorf("adding groups: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Added group(s) to user %d\n", userID)
			return nil
		},
	}

	cmd.Flags().StringVar(&groupIDsFlag, "group-ids", "", "Comma-separated group IDs")
	cmd.Flags().Int64Var(&revision, "revision", 0, "Explicit revision (skips auto-fetch)")
	return cmd
}

func newUsersGroupsRemoveCmd(state *appState) *cobra.Command {
	var (
		groupIDsFlag string
		revision     int64
	)

	cmd := &cobra.Command{
		Use:   "remove <userId> [groupId]",
		Short: "Remove groups from an identity user",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid userId %q: %w", args[0], err)
			}

			var groupIDs []int
			switch {
			case len(args) == 2:
				n, err := strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid groupId %q: %w", args[1], err)
				}
				groupIDs = []int{n}
			case groupIDsFlag != "":
				var parseErr error
				groupIDs, parseErr = api.ParseGroupIDs(groupIDsFlag)
				if parseErr != nil {
					return parseErr
				}
			default:
				return fmt.Errorf("provide a groupId argument or --group-ids flag")
			}

			if revision == 0 {
				current, err := state.client.GetUser(cmd.Context(), userID)
				if err != nil {
					return fmt.Errorf("fetching current revision: %w", err)
				}
				revision = current.Revision
			}

			if _, err := state.client.RemoveUserGroups(cmd.Context(), userID, groupIDs, revision); err != nil {
				return fmt.Errorf("removing groups: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Removed group(s) from user %d\n", userID)
			return nil
		},
	}

	cmd.Flags().StringVar(&groupIDsFlag, "group-ids", "", "Comma-separated group IDs")
	cmd.Flags().Int64Var(&revision, "revision", 0, "Explicit revision (skips auto-fetch)")
	return cmd
}

// ---------------------------------------------------------------------------
// users bulk
// ---------------------------------------------------------------------------

func newUsersBulkCmd(state *appState) *cobra.Command {
	bulk := &cobra.Command{
		Use:   "bulk",
		Short: "Apply a bulk action to multiple identity users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	for _, action := range []struct {
		use    string
		short  string
		apiKey string
	}{
		{"start-tracking", "Enable tracking for users", "StartTracking"},
		{"stop-tracking", "Disable tracking for users", "StopTracking"},
		{"delete-entity", "Permanently delete user entity records", "DeleteEntity"},
		{"delete-data", "Delete recorded activity data for users", "DeleteData"},
	} {
		bulk.AddCommand(newBulkActionCmd(state, action.use, action.short, action.apiKey))
	}

	return bulk
}

func newBulkActionCmd(state *appState, use, short, apiAction string) *cobra.Command {
	var (
		idsFlag   string
		asJSON    bool
		fromStdin bool
	)

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			var req api.BulkActionRequest

			if fromStdin {
				if err := stdin.EnsurePiped(); err != nil {
					return err
				}
				var err error
				req, err = stdin.ReadJSON[api.BulkActionRequest](os.Stdin)
				if err != nil {
					return err
				}
				if len(req.Actions) == 0 || len(req.Data) == 0 {
					return fmt.Errorf("--from-stdin: payload must include non-empty \"actions\" and \"data\" fields")
				}
				if req.Actions[0] != apiAction {
					return fmt.Errorf("--from-stdin: actions mismatch: expected [%s]", apiAction)
				}
			} else {
				if idsFlag == "" {
					return fmt.Errorf("--ids is required")
				}

				parts := strings.Split(idsFlag, ",")
				ids := make([]int64, 0, len(parts))
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					id, err := strconv.ParseInt(p, 10, 64)
					if err != nil {
						return fmt.Errorf("invalid id %q: %w", p, err)
					}
					ids = append(ids, id)
				}
				if len(ids) == 0 {
					return fmt.Errorf("--ids must contain at least one entity ID")
				}

				revisions, errs := state.client.FetchRevisions(cmd.Context(), ids, 10)
				for _, e := range errs {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", e)
				}

				data := make([]api.BulkEntityData, 0, len(revisions))
				for _, id := range ids {
					rev, ok := revisions[id]
					if !ok {
						continue
					}
					data = append(data, api.BulkEntityData{
						EntityID: int(id),
						Revision: int(rev),
					})
				}

				if len(data) == 0 {
					return fmt.Errorf("no entities could be fetched for bulk action")
				}

				req = api.BulkActionRequest{
					Actions: []string{apiAction},
					Data:    data,
				}
			}

			result, err := state.client.BulkAction(cmd.Context(), req)
			if err != nil {
				return fmt.Errorf("bulk action: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), result)
			}

			rows := make([][]string, 0, len(result.Successful)+len(result.Failures))
			for _, s := range result.Successful {
				rows = append(rows, []string{strconv.Itoa(s.EntityID), "ok"})
			}
			for _, f := range result.Failures {
				rows = append(rows, []string{strconv.Itoa(f.EntityID), "failed: " + f.Message})
			}
			output.Table(cmd.OutOrStdout(), []string{"ENTITY ID", "RESULT"}, rows)

			if len(result.Failures) > 0 {
				return fmt.Errorf("%d entity(ies) failed", len(result.Failures))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated entity IDs")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output BulkActionResponse JSON")
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read full BulkActionRequest JSON from stdin instead of --ids")
	return cmd
}

// ---------------------------------------------------------------------------
// Formatting helpers
// ---------------------------------------------------------------------------

func identityGroupNames(groups []api.IdentityGroupRef) string {
	if len(groups) == 0 {
		return ""
	}
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.GroupName
	}
	return strings.Join(names, ", ")
}

func identityFieldList(fields []api.IdentityField) string {
	if len(fields) == 0 {
		return ""
	}
	vals := make([]string, len(fields))
	for i, f := range fields {
		vals[i] = f.Value
	}
	return strings.Join(vals, ", ")
}

func identityAgentSummary(agents []api.IdentityAgent) string {
	if len(agents) == 0 {
		return ""
	}
	parts := make([]string, len(agents))
	for i, a := range agents {
		parts[i] = a.UserName + " (" + a.LicenseStatus + ")"
	}
	return strings.Join(parts, ", ")
}
