package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/output"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)

// newSchedulesCmd returns the "schedules" resource command with all subcommands wired.
func newSchedulesCmd(state *appState) *cobra.Command {
	schedules := &cobra.Command{
		Use:   "schedules",
		Short: "Manage ActivTrak schedules",
		Long:  `Commands for listing, inspecting, and managing work schedules and schedule assignments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	schedules.AddCommand(newSchedulesListCmd(state))
	schedules.AddCommand(newSchedulesGetCmd(state))
	schedules.AddCommand(newSchedulesCreateCmd(state))
	schedules.AddCommand(newSchedulesDeleteCmd(state))
	schedules.AddCommand(newSchedulesReportingCmd(state))
	schedules.AddCommand(newSchedulesShiftCmd(state))
	schedules.AddCommand(newSchedulesUsersCmd(state))
	schedules.AddCommand(newSchedulesUserCmd(state))

	return schedules
}

// newSchedulesListCmd implements "schedules list".
func newSchedulesListCmd(state *appState) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all schedules",
		RunE: func(cmd *cobra.Command, args []string) error {
			scheds, err := state.client.ListSchedules(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing schedules: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), scheds)
			}

			rows := make([][]string, len(scheds))
			for i, s := range scheds {
				rows[i] = []string{
					strconv.Itoa(s.ID),
					s.Name,
					s.Type,
					strconv.FormatBool(s.IsDefault),
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "NAME", "TYPE", "DEFAULT"}, rows)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	return cmd
}

// newSchedulesGetCmd implements "schedules get <id>".
func newSchedulesGetCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a schedule by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid schedule ID %q: must be an integer", args[0])
			}
			s, err := state.client.GetSchedule(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("getting schedule %d: %w", id, err)
			}
			output.KeyValue(cmd.OutOrStdout(), map[string]string{
				"id":        strconv.Itoa(s.ID),
				"name":      s.Name,
				"type":      s.Type,
				"isDefault": strconv.FormatBool(s.IsDefault),
			})
			return nil
		},
	}
}

// newSchedulesCreateCmd implements "schedules create --file <path>".
func newSchedulesCreateCmd(state *appState) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a schedule from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			id, err := state.client.CreateSchedule(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("creating schedule: %w", err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), strconv.Itoa(id))
			return nil
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON file (required)")
	return cmd
}

// newSchedulesDeleteCmd implements "schedules delete <id>".
func newSchedulesDeleteCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a schedule by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid schedule ID %q: must be an integer", args[0])
			}
			if err := state.client.DeleteSchedule(cmd.Context(), id); err != nil {
				return fmt.Errorf("deleting schedule %d: %w", id, err)
			}
			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Deleted schedule %d\n", id)
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), strconv.Itoa(id))
			}
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// schedules reporting
// ---------------------------------------------------------------------------

func newSchedulesReportingCmd(state *appState) *cobra.Command {
	reporting := &cobra.Command{
		Use:   "reporting",
		Short: "Manage reporting schedule defaults and users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	reporting.AddCommand(newSchedulesReportingDefaultCmd(state))
	reporting.AddCommand(newSchedulesReportingUsersCmd(state))
	return reporting
}

func newSchedulesReportingDefaultCmd(state *appState) *cobra.Command {
	def := &cobra.Command{
		Use:   "default",
		Short: "Manage the default reporting schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	def.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get the default reporting schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := state.client.GetReportingDefault(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting reporting default: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), map[string]string{
				"id":        strconv.Itoa(s.ID),
				"name":      s.Name,
				"type":      s.Type,
				"isDefault": strconv.FormatBool(s.IsDefault),
			})
			return nil
		},
	})
	def.AddCommand(&cobra.Command{
		Use:   "set <id>",
		Short: "Set the default reporting schedule by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid schedule ID %q: must be an integer", args[0])
			}
			if err := state.client.SetReportingDefault(cmd.Context(), id); err != nil {
				return fmt.Errorf("setting reporting default: %w", err)
			}
			printSuccess(cmd, fmt.Sprintf("Reporting default set to schedule %d", id))
			return nil
		},
	})
	return def
}

func newSchedulesReportingUsersCmd(state *appState) *cobra.Command {
	users := &cobra.Command{
		Use:   "users",
		Short: "Manage reporting schedule users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	users.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List users in the reporting schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			us, err := state.client.GetReportingUsers(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing reporting users: %w", err)
			}
			rows := make([][]string, len(us))
			for i, u := range us {
				rows[i] = []string{strconv.Itoa(u.UserID), strconv.Itoa(u.ScheduleID), u.Name}
			}
			output.Table(cmd.OutOrStdout(), []string{"USER ID", "SCHEDULE ID", "NAME"}, rows)
			return nil
		},
	})

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove users from the reporting schedule by IDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			idsFlag, _ := cmd.Flags().GetString("ids")
			if idsFlag == "" {
				return fmt.Errorf("--ids is required")
			}
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if err := state.client.RemoveReportingUsers(cmd.Context(), ids); err != nil {
				return fmt.Errorf("removing reporting users: %w", err)
			}
			printSuccess(cmd, fmt.Sprintf("Removed %d users from reporting schedule", len(ids)))
			return nil
		},
	}
	removeCmd.Flags().String("ids", "", "Comma-separated user IDs to remove (required)")
	users.AddCommand(removeCmd)
	return users
}

// ---------------------------------------------------------------------------
// schedules shift
// ---------------------------------------------------------------------------

func newSchedulesShiftCmd(state *appState) *cobra.Command {
	shift := &cobra.Command{
		Use:   "shift",
		Short: "Manage shift schedule defaults and users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	shift.AddCommand(newSchedulesShiftDefaultCmd(state))
	shift.AddCommand(newSchedulesShiftUsersCmd(state))
	return shift
}

func newSchedulesShiftDefaultCmd(state *appState) *cobra.Command {
	def := &cobra.Command{
		Use:   "default",
		Short: "Manage the default shift schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	def.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get the default shift schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := state.client.GetShiftDefault(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting shift default: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), map[string]string{
				"id":        strconv.Itoa(s.ID),
				"name":      s.Name,
				"type":      s.Type,
				"isDefault": strconv.FormatBool(s.IsDefault),
			})
			return nil
		},
	})
	def.AddCommand(&cobra.Command{
		Use:   "set <id>",
		Short: "Set the default shift schedule by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid schedule ID %q: must be an integer", args[0])
			}
			if err := state.client.SetShiftDefault(cmd.Context(), id); err != nil {
				return fmt.Errorf("setting shift default: %w", err)
			}
			printSuccess(cmd, fmt.Sprintf("Shift default set to schedule %d", id))
			return nil
		},
	})
	return def
}

func newSchedulesShiftUsersCmd(state *appState) *cobra.Command {
	users := &cobra.Command{
		Use:   "users",
		Short: "Manage shift schedule users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	users.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List users in the shift schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			us, err := state.client.GetShiftUsers(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing shift users: %w", err)
			}
			rows := make([][]string, len(us))
			for i, u := range us {
				rows[i] = []string{strconv.Itoa(u.UserID), strconv.Itoa(u.ScheduleID), u.Name}
			}
			output.Table(cmd.OutOrStdout(), []string{"USER ID", "SCHEDULE ID", "NAME"}, rows)
			return nil
		},
	})

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove users from the shift schedule by IDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			idsFlag, _ := cmd.Flags().GetString("ids")
			if idsFlag == "" {
				return fmt.Errorf("--ids is required")
			}
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if err := state.client.RemoveShiftUsers(cmd.Context(), ids); err != nil {
				return fmt.Errorf("removing shift users: %w", err)
			}
			printSuccess(cmd, fmt.Sprintf("Removed %d users from shift schedule", len(ids)))
			return nil
		},
	}
	removeCmd.Flags().String("ids", "", "Comma-separated user IDs to remove (required)")
	users.AddCommand(removeCmd)
	return users
}

// ---------------------------------------------------------------------------
// schedules users <schedule-id>
// ---------------------------------------------------------------------------

func newSchedulesUsersCmd(state *appState) *cobra.Command {
	users := &cobra.Command{
		Use:   "users",
		Short: "Manage schedule user assignments",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	listCmd := &cobra.Command{
		Use:   "list <schedule-id>",
		Short: "List users assigned to a schedule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scheduleID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid schedule ID %q: must be an integer", args[0])
			}
			us, err := state.client.GetScheduleUsers(cmd.Context(), scheduleID)
			if err != nil {
				return fmt.Errorf("listing users for schedule %d: %w", scheduleID, err)
			}
			rows := make([][]string, len(us))
			for i, u := range us {
				rows[i] = []string{strconv.Itoa(u.UserID), strconv.Itoa(u.ScheduleID), u.Name}
			}
			output.Table(cmd.OutOrStdout(), []string{"USER ID", "SCHEDULE ID", "NAME"}, rows)
			return nil
		},
	}
	users.AddCommand(listCmd)

	setCmd := &cobra.Command{
		Use:   "set <schedule-id>",
		Short: "Set users assigned to a schedule (replaces current assignment)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scheduleID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid schedule ID %q: must be an integer", args[0])
			}
			idsFlag, _ := cmd.Flags().GetString("ids")
			if idsFlag == "" {
				return fmt.Errorf("--ids is required")
			}
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if err := state.client.SetScheduleUsers(cmd.Context(), scheduleID, ids); err != nil {
				return fmt.Errorf("setting users for schedule %d: %w", scheduleID, err)
			}
			printSuccess(cmd, fmt.Sprintf("Updated %d users for schedule %d", len(ids), scheduleID))
			return nil
		},
	}
	setCmd.Flags().String("ids", "", "Comma-separated user IDs to assign (required)")
	users.AddCommand(setCmd)

	return users
}

// ---------------------------------------------------------------------------
// schedules user
// ---------------------------------------------------------------------------

func newSchedulesUserCmd(state *appState) *cobra.Command {
	user := &cobra.Command{
		Use:   "user",
		Short: "Manage individual user schedule assignments",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	user.AddCommand(newSchedulesUserMoveCmd(state))
	user.AddCommand(newSchedulesUserGetCmd(state))
	user.AddCommand(newSchedulesUserRemoveCmd(state))
	return user
}

// newSchedulesUserMoveCmd implements "schedules user move --schedule <id> --user <id>".
func newSchedulesUserMoveCmd(state *appState) *cobra.Command {
	var (
		scheduleID int
		userID     int
	)

	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move a user to a specific schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			if scheduleID == 0 {
				return fmt.Errorf("--schedule is required")
			}
			if userID == 0 {
				return fmt.Errorf("--user is required")
			}
			if err := state.client.MoveUserToSchedule(cmd.Context(), scheduleID, userID); err != nil {
				return fmt.Errorf("moving user %d to schedule %d: %w", userID, scheduleID, err)
			}
			printSuccess(cmd, fmt.Sprintf("User %d moved to schedule %d", userID, scheduleID))
			return nil
		},
	}
	cmd.Flags().IntVar(&scheduleID, "schedule", 0, "Schedule ID (required)")
	cmd.Flags().IntVar(&userID, "user", 0, "User ID (required)")
	return cmd
}

// newSchedulesUserGetCmd implements "schedules user get <user-id> --type reporting|shift".
func newSchedulesUserGetCmd(state *appState) *cobra.Command {
	var schedType string

	cmd := &cobra.Command{
		Use:   "get <user-id>",
		Short: "Get a user's reporting or shift schedule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid user ID %q: must be an integer", args[0])
			}

			var s interface {
				// Use interface to avoid duplicating call logic
			}
			_ = s

			switch schedType {
			case "reporting":
				sched, err := state.client.GetUserReportingSchedule(cmd.Context(), userID)
				if err != nil {
					return fmt.Errorf("getting reporting schedule for user %d: %w", userID, err)
				}
				output.KeyValue(cmd.OutOrStdout(), map[string]string{
					"id":        strconv.Itoa(sched.ID),
					"name":      sched.Name,
					"type":      sched.Type,
					"isDefault": strconv.FormatBool(sched.IsDefault),
				})
			case "shift":
				sched, err := state.client.GetUserShiftSchedule(cmd.Context(), userID)
				if err != nil {
					return fmt.Errorf("getting shift schedule for user %d: %w", userID, err)
				}
				output.KeyValue(cmd.OutOrStdout(), map[string]string{
					"id":        strconv.Itoa(sched.ID),
					"name":      sched.Name,
					"type":      sched.Type,
					"isDefault": strconv.FormatBool(sched.IsDefault),
				})
			default:
				return fmt.Errorf("--type must be 'reporting' or 'shift'")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&schedType, "type", "", "Schedule type: reporting or shift (required)")
	return cmd
}

// newSchedulesUserRemoveCmd implements "schedules user remove <user-id> --type reporting|shift".
func newSchedulesUserRemoveCmd(state *appState) *cobra.Command {
	var schedType string

	cmd := &cobra.Command{
		Use:   "remove <user-id>",
		Short: "Remove a user from their reporting or shift schedule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid user ID %q: must be an integer", args[0])
			}
			switch schedType {
			case "reporting":
				if err := state.client.RemoveUserFromReportingSchedules(cmd.Context(), userID); err != nil {
					return fmt.Errorf("removing user %d from reporting schedules: %w", userID, err)
				}
				printSuccess(cmd, fmt.Sprintf("User %d removed from reporting schedules", userID))
			case "shift":
				if err := state.client.RemoveUserFromShiftSchedules(cmd.Context(), userID); err != nil {
					return fmt.Errorf("removing user %d from shift schedules: %w", userID, err)
				}
				printSuccess(cmd, fmt.Sprintf("User %d removed from shift schedules", userID))
			default:
				return fmt.Errorf("--type must be 'reporting' or 'shift'")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&schedType, "type", "", "Schedule type: reporting or shift (required)")
	return cmd
}
