package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/bulk"
	"github.com/activtrak-mfinlayson/atadmin/internal/output"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)

// newGroupsCmd returns the "groups" resource command with all subcommands wired.
func newGroupsCmd(state *appState) *cobra.Command {
	groups := &cobra.Command{
		Use:   "groups",
		Short: "Manage ActivTrak groups",
		Long:  `Commands for listing, inspecting, and managing user groups.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	groups.AddCommand(newGroupsListCmd(state))
	groups.AddCommand(newGroupsSummaryCmd(state))
	groups.AddCommand(newGroupsGetCmd(state))
	groups.AddCommand(newGroupsSearchCmd(state))
	groups.AddCommand(newGroupsCreateCmd(state))
	groups.AddCommand(newGroupsRenameCmd(state))
	groups.AddCommand(newGroupsDeleteCmd(state))
	groups.AddCommand(newGroupsMembersCmd(state))
	groups.AddCommand(newGroupsClientsCmd(state))
	groups.AddCommand(newGroupsDevicesCmd(state))
	groups.AddCommand(newGroupsMembershipCmd(state))

	return groups
}

// newGroupsListCmd implements "groups list".
func newGroupsListCmd(state *appState) *cobra.Command {
	var (
		page     int
		pageSize int
		asJSON   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			groups, err := state.client.ListGroups(cmd.Context(), page, pageSize)
			if err != nil {
				return fmt.Errorf("listing groups: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), groups)
			}

			rows := make([][]string, len(groups))
			for i, g := range groups {
				rows[i] = []string{
					strconv.Itoa(g.ID),
					g.Name,
					strconv.Itoa(g.MemberCount),
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "NAME", "MEMBER COUNT"}, rows)
			return nil
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 50, "Number of results per page")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")

	return cmd
}

// newGroupsSummaryCmd implements "groups summary".
func newGroupsSummaryCmd(state *appState) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Get a summary of all groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			groups, err := state.client.GetGroupSummary(cmd.Context())
			if err != nil {
				return fmt.Errorf("fetching groups summary: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), groups)
			}

			rows := make([][]string, len(groups))
			for i, g := range groups {
				rows[i] = []string{
					strconv.Itoa(g.ID),
					g.Name,
					strconv.Itoa(g.MemberCount),
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "NAME", "MEMBER COUNT"}, rows)
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")

	return cmd
}

// newGroupsGetCmd implements "groups get <id>".
func newGroupsGetCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a group by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid group ID %q: must be an integer", args[0])
			}

			group, err := state.client.GetGroup(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("getting group %d: %w", id, err)
			}

			output.KeyValue(cmd.OutOrStdout(), map[string]string{
				"id":          strconv.Itoa(group.ID),
				"name":        group.Name,
				"memberCount": strconv.Itoa(group.MemberCount),
			})
			return nil
		},
	}
}

// newGroupsSearchCmd implements "groups search <prefix>".
func newGroupsSearchCmd(state *appState) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "search <prefix>",
		Short: "Search groups by name prefix",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groups, err := state.client.SearchGroups(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("searching groups: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), groups)
			}

			rows := make([][]string, len(groups))
			for i, g := range groups {
				rows[i] = []string{
					strconv.Itoa(g.ID),
					g.Name,
					strconv.Itoa(g.MemberCount),
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "NAME", "MEMBER COUNT"}, rows)
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")

	return cmd
}

// newGroupsCreateCmd implements "groups create <name>".
func newGroupsCreateCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := state.client.CreateGroup(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("creating group %q: %w", args[0], err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), strconv.Itoa(id))
			return nil
		},
	}
}

// newGroupsRenameCmd implements "groups rename <id> --name <str>".
func newGroupsRenameCmd(state *appState) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "rename <id>",
		Short: "Rename a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid group ID %q: must be an integer", args[0])
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			if err := state.client.RenameGroup(cmd.Context(), id, name); err != nil {
				return fmt.Errorf("renaming group %d: %w", id, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Renamed group %d to %q\n", id, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New group name (required)")

	return cmd
}

// newGroupsDeleteCmd implements "groups delete --ids <csv>".
func newGroupsDeleteCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete groups by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.DeleteGroups(cmd.Context(), ids); err != nil {
				return fmt.Errorf("deleting groups: %w", err)
			}

			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted %d groups\n", len(ids))
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "deleted")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated group IDs to delete (required)")

	return cmd
}

// ---------------------------------------------------------------------------
// groups members subcommand group
// ---------------------------------------------------------------------------

// newGroupsMembersCmd returns the "groups members" subcommand group.
func newGroupsMembersCmd(state *appState) *cobra.Command {
	members := &cobra.Command{
		Use:   "members",
		Short: "Manage group members",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	members.AddCommand(newGroupsMembersListCmd(state))
	members.AddCommand(newGroupsMembersAddCmd(state))
	members.AddCommand(newGroupsMembersRemoveCmd(state))
	members.AddCommand(newGroupsMembersExportCmd(state))
	members.AddCommand(newGroupsMembersImportCmd(state))

	return members
}

// newGroupsMembersListCmd implements "groups members list [--group <id>]".
func newGroupsMembersListCmd(state *appState) *cobra.Command {
	var (
		groupID  int
		page     int
		pageSize int
		asJSON   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List group members",
		RunE: func(cmd *cobra.Command, args []string) error {
			if groupID != 0 {
				members, err := state.client.ListGroupMembers(cmd.Context(), groupID)
				if err != nil {
					return fmt.Errorf("listing members for group %d: %w", groupID, err)
				}

				if asJSON {
					return output.JSON(cmd.OutOrStdout(), members)
				}

				rows := make([][]string, len(members))
				for i, m := range members {
					rows[i] = []string{strconv.Itoa(m.MemberID), m.MemberName, m.MemberAlias, m.MemberType}
				}
				output.Table(cmd.OutOrStdout(), []string{"ID", "NAME", "ALIAS", "TYPE"}, rows)
				return nil
			}

			members, err := state.client.ListMembers(cmd.Context(), page, pageSize)
			if err != nil {
				return fmt.Errorf("listing members: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), members)
			}

			rows := make([][]string, len(members))
			for i, m := range members {
				rows[i] = []string{strconv.Itoa(m.GroupID), strconv.Itoa(m.MemberID), m.MemberName, m.MemberType}
			}
			output.Table(cmd.OutOrStdout(), []string{"GROUP ID", "ID", "NAME", "TYPE"}, rows)
			return nil
		},
	}

	cmd.Flags().IntVar(&groupID, "group", 0, "Filter by group ID")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 50, "Number of results per page")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")

	return cmd
}

// newGroupsMembersAddCmd implements "groups members add --group <id> --member <id> --type <str>".
func newGroupsMembersAddCmd(state *appState) *cobra.Command {
	var (
		groupID    int
		memberID   int
		memberType string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a member to a group",
		RunE: func(cmd *cobra.Command, args []string) error {
			if groupID == 0 {
				return fmt.Errorf("--group is required")
			}
			if memberID == 0 {
				return fmt.Errorf("--member is required")
			}
			if memberType == "" {
				return fmt.Errorf("--type is required")
			}
			if memberType != "client" && memberType != "device" {
				return fmt.Errorf("--type must be 'client' or 'device', got %q", memberType)
			}

			if err := state.client.AddMembers(cmd.Context(), groupID, memberID, memberType); err != nil {
				return fmt.Errorf("adding member %d to group %d: %w", memberID, groupID, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added member %d (%s) to group %d\n", memberID, memberType, groupID)
			return nil
		},
	}

	cmd.Flags().IntVar(&groupID, "group", 0, "Group ID (required)")
	cmd.Flags().IntVar(&memberID, "member", 0, "Member ID (required)")
	cmd.Flags().StringVar(&memberType, "type", "", "Member type: client or device (required)")

	return cmd
}

// newGroupsMembersRemoveCmd implements "groups members remove --group <id> --member <id>".
func newGroupsMembersRemoveCmd(state *appState) *cobra.Command {
	var (
		groupID  int
		memberID int
	)

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a member from a group",
		RunE: func(cmd *cobra.Command, args []string) error {
			if groupID == 0 {
				return fmt.Errorf("--group is required")
			}
			if memberID == 0 {
				return fmt.Errorf("--member is required")
			}

			if err := state.client.RemoveMembers(cmd.Context(), groupID, memberID); err != nil {
				return fmt.Errorf("removing member %d from group %d: %w", memberID, groupID, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed member %d from group %d\n", memberID, groupID)
			return nil
		},
	}

	cmd.Flags().IntVar(&groupID, "group", 0, "Group ID (required)")
	cmd.Flags().IntVar(&memberID, "member", 0, "Member ID (required)")

	return cmd
}

// newGroupsMembersExportCmd implements "groups members export [--output <path>]".
func newGroupsMembersExportCmd(state *appState) *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export group members to a file or stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := state.client.ExportMembers(cmd.Context())
			if err != nil {
				return fmt.Errorf("exporting members: %w", err)
			}

			if outputPath == "" {
				_, err = cmd.OutOrStdout().Write(data)
				return err
			}

			if err := os.WriteFile(outputPath, data, 0644); err != nil {
				return fmt.Errorf("writing export to %q: %w", outputPath, err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Exported members to %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&outputPath, "output", "", "Output file path (defaults to stdout)")

	return cmd
}

// newGroupsMembersImportCmd implements "groups members import --file <path>".
func newGroupsMembersImportCmd(state *appState) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import group members from a JSON or CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			records, err := bulk.ParseFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}

			if err := state.client.ImportMembers(cmd.Context(), records); err != nil {
				return fmt.Errorf("importing members: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Imported %d member records\n", len(records))
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON or CSV file (required)")

	return cmd
}

// ---------------------------------------------------------------------------
// groups clients subcommand group
// ---------------------------------------------------------------------------

// newGroupsClientsCmd returns the "groups clients" subcommand group.
func newGroupsClientsCmd(state *appState) *cobra.Command {
	clients := &cobra.Command{
		Use:   "clients",
		Short: "Manage clients within a group",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	clients.AddCommand(newGroupsClientsAddCmd(state))
	clients.AddCommand(newGroupsClientsRemoveCmd(state))

	return clients
}

// newGroupsClientsAddCmd implements "groups clients add <group-id> --ids <csv>".
func newGroupsClientsAddCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "add <group-id>",
		Short: "Add clients to a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid group ID %q: must be an integer", args[0])
			}

			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.AddClientsToGroup(cmd.Context(), groupID, ids); err != nil {
				return fmt.Errorf("adding clients to group %d: %w", groupID, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added %d clients to group %d\n", len(ids), groupID)
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated client IDs to add (required)")

	return cmd
}

// newGroupsClientsRemoveCmd implements "groups clients remove <group-id> --ids <csv>".
func newGroupsClientsRemoveCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "remove <group-id>",
		Short: "Remove clients from a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid group ID %q: must be an integer", args[0])
			}

			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.RemoveClientsFromGroup(cmd.Context(), groupID, ids); err != nil {
				return fmt.Errorf("removing clients from group %d: %w", groupID, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed %d clients from group %d\n", len(ids), groupID)
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated client IDs to remove (required)")

	return cmd
}

// ---------------------------------------------------------------------------
// groups devices subcommand group
// ---------------------------------------------------------------------------

// newGroupsDevicesCmd returns the "groups devices" subcommand group.
func newGroupsDevicesCmd(state *appState) *cobra.Command {
	devices := &cobra.Command{
		Use:   "devices",
		Short: "Manage devices within a group",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	devices.AddCommand(newGroupsDevicesAddCmd(state))
	devices.AddCommand(newGroupsDevicesRemoveCmd(state))

	return devices
}

// newGroupsDevicesAddCmd implements "groups devices add <group-id> --ids <csv>".
func newGroupsDevicesAddCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "add <group-id>",
		Short: "Add devices to a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid group ID %q: must be an integer", args[0])
			}

			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.AddDevicesToGroup(cmd.Context(), groupID, ids); err != nil {
				return fmt.Errorf("adding devices to group %d: %w", groupID, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added %d devices to group %d\n", len(ids), groupID)
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated device IDs to add (required)")

	return cmd
}

// newGroupsDevicesRemoveCmd implements "groups devices remove <group-id> --ids <csv>".
func newGroupsDevicesRemoveCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "remove <group-id>",
		Short: "Remove devices from a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid group ID %q: must be an integer", args[0])
			}

			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.RemoveDevicesFromGroup(cmd.Context(), groupID, ids); err != nil {
				return fmt.Errorf("removing devices from group %d: %w", groupID, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed %d devices from group %d\n", len(ids), groupID)
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated device IDs to remove (required)")

	return cmd
}

// newGroupsMembershipCmd returns the "groups membership" subcommand.
func newGroupsMembershipCmd(state *appState) *cobra.Command {
	membership := &cobra.Command{
		Use:   "membership",
		Short: "Look up a specific group membership",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	membership.AddCommand(newGroupsMembershipGetCmd(state))
	return membership
}

// newGroupsMembershipGetCmd implements "groups membership get <group-id> <type> <member-id>".
func newGroupsMembershipGetCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "get <group-id> <type> <member-id>",
		Short: "Get a specific member within a group",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid group ID %q: must be an integer", args[0])
			}
			memberType := args[1]
			memberID, err := strconv.Atoi(args[2])
			if err != nil {
				return fmt.Errorf("invalid member ID %q: must be an integer", args[2])
			}

			m, err := state.client.GetGroupMember(cmd.Context(), groupID, memberType, memberID)
			if err != nil {
				return fmt.Errorf("getting group member: %w", err)
			}

			output.KeyValue(cmd.OutOrStdout(), map[string]string{
				"groupId":    strconv.Itoa(m.GroupID),
				"memberId":   strconv.Itoa(m.MemberID),
				"memberType": m.MemberType,
			})
			return nil
		},
	}
}
