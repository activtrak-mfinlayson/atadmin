package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/output"
	"github.com/activtrak-mfinlayson/atadmin/internal/stdin"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)

// newAlarmsCmd returns the "alarms" resource command with all subcommands wired.
func newAlarmsCmd(state *appState) *cobra.Command {
	alarms := &cobra.Command{
		Use:   "alarms",
		Short: "Manage ActivTrak alarms",
		Long:  `Commands for listing, inspecting, creating, updating, and deleting alarms.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	alarms.AddCommand(newAlarmsListCmd(state))
	alarms.AddCommand(newAlarmsGetCmd(state))
	alarms.AddCommand(newAlarmsDetailsCmd(state))
	alarms.AddCommand(newAlarmsCreateCmd(state))
	alarms.AddCommand(newAlarmsUpdateCmd(state))
	alarms.AddCommand(newAlarmsDeleteCmd(state))
	alarms.AddCommand(newAlarmsConditionsCmd(state))
	alarms.AddCommand(newAlarmsFieldsCmd(state))

	return alarms
}

// newAlarmsListCmd implements "alarms list".
func newAlarmsListCmd(state *appState) *cobra.Command {
	var (
		page        int
		pageSize    int
		asJSON      bool
		fieldsFlag  string
		summaryFlag bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List alarms",
		RunE: func(cmd *cobra.Command, args []string) error {
			alarms, err := state.client.ListAlarms(cmd.Context(), page, pageSize)
			if err != nil {
				return fmt.Errorf("listing alarms: %w", err)
			}

			if asJSON {
				if summaryFlag {
					return output.JSONSummary(cmd.OutOrStdout(), len(alarms), nil, len(alarms) == pageSize)
				}
				if fieldsFlag != "" {
					generic, err := output.ToGeneric(alarms)
					if err != nil {
						return fmt.Errorf("serializing results: %w", err)
					}
					return output.JSON(cmd.OutOrStdout(), output.FilterFields(generic, strings.Split(fieldsFlag, ",")))
				}
				return output.JSON(cmd.OutOrStdout(), alarms)
			}

			rows := make([][]string, len(alarms))
			for i, a := range alarms {
				rows[i] = []string{
					strconv.Itoa(a.ID),
					a.Name,
					a.Type,
					strconv.FormatBool(a.Enabled),
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "NAME", "TYPE", "ENABLED"}, rows)
			return nil
		},
	}
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 50, "Number of results per page")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	cmd.Flags().StringVar(&fieldsFlag, "fields", "", "Comma-separated top-level JSON keys to include (e.g. id,name)")
	cmd.Flags().BoolVar(&summaryFlag, "summary", false, "Return aggregate statistics instead of full results")
	return cmd
}

// newAlarmsGetCmd implements "alarms get <id>".
func newAlarmsGetCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an alarm by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid alarm ID %q: must be an integer", args[0])
			}
			alarm, err := state.client.GetAlarm(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("getting alarm %d: %w", id, err)
			}
			output.KeyValue(cmd.OutOrStdout(), map[string]string{
				"id":      strconv.Itoa(alarm.ID),
				"name":    alarm.Name,
				"type":    alarm.Type,
				"enabled": strconv.FormatBool(alarm.Enabled),
			})
			return nil
		},
	}
}

// newAlarmsDetailsCmd implements "alarms details <id>".
func newAlarmsDetailsCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "details <id>",
		Short: "Get detailed alarm configuration by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid alarm ID %q: must be an integer", args[0])
			}
			details, err := state.client.GetAlarmDetails(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("getting alarm details %d: %w", id, err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(details))
			return nil
		},
	}
}

// newAlarmsCreateCmd implements "alarms create --file <path>".
func newAlarmsCreateCmd(state *appState) *cobra.Command {
	var (
		filePath  string
		fromStdin bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an alarm from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath != "" && fromStdin {
				return fmt.Errorf("--from-stdin: --file and --from-stdin are mutually exclusive")
			}

			var body map[string]any
			var err error

			if fromStdin {
				if err := stdin.EnsurePiped(); err != nil {
					return err
				}
				body, err = stdin.ReadJSON[map[string]any](os.Stdin)
				if err != nil {
					return err
				}
			} else {
				if filePath == "" {
					return fmt.Errorf("--file is required")
				}
				body, err = readJSONObjectFile(filePath)
				if err != nil {
					return fmt.Errorf("reading file %q: %w", filePath, err)
				}
			}

			if err := state.client.SaveAlarms(cmd.Context(), body); err != nil {
				return fmt.Errorf("creating alarm: %w", err)
			}
			printSuccess(cmd, "Alarm created")
			return nil
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON file")
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read JSON object from stdin instead of --file")
	return cmd
}

// newAlarmsUpdateCmd implements "alarms update --file <path>".
func newAlarmsUpdateCmd(state *appState) *cobra.Command {
	var (
		filePath  string
		fromStdin bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an alarm from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath != "" && fromStdin {
				return fmt.Errorf("--from-stdin: --file and --from-stdin are mutually exclusive")
			}

			var body map[string]any
			var err error

			if fromStdin {
				if err := stdin.EnsurePiped(); err != nil {
					return err
				}
				body, err = stdin.ReadJSON[map[string]any](os.Stdin)
				if err != nil {
					return err
				}
			} else {
				if filePath == "" {
					return fmt.Errorf("--file is required")
				}
				body, err = readJSONObjectFile(filePath)
				if err != nil {
					return fmt.Errorf("reading file %q: %w", filePath, err)
				}
			}

			if err := state.client.SaveAlarm(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating alarm: %w", err)
			}
			printSuccess(cmd, "Alarm updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON file")
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read JSON object from stdin instead of --file")
	return cmd
}

// newAlarmsDeleteCmd implements "alarms delete <id>".
func newAlarmsDeleteCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an alarm by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid alarm ID %q: must be an integer", args[0])
			}
			if err := state.client.DeleteAlarm(cmd.Context(), id); err != nil {
				return fmt.Errorf("deleting alarm %d: %w", id, err)
			}
			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Deleted alarm %d\n", id)
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), strconv.Itoa(id))
			}
			return nil
		},
	}
}

// newAlarmsConditionsCmd implements "alarms conditions".
func newAlarmsConditionsCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "conditions",
		Short: "List available alarm condition types",
		RunE: func(cmd *cobra.Command, args []string) error {
			conditions, err := state.client.GetAlarmConditions(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting alarm conditions: %w", err)
			}
			return output.JSON(cmd.OutOrStdout(), conditions)
		},
	}
}

// newAlarmsFieldsCmd implements "alarms fields".
func newAlarmsFieldsCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "fields",
		Short: "List available alarm field types",
		RunE: func(cmd *cobra.Command, args []string) error {
			fields, err := state.client.GetAlarmFields(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting alarm fields: %w", err)
			}
			return output.JSON(cmd.OutOrStdout(), fields)
		},
	}
}
