package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/output"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)

// newDevicesCmd returns the "devices" resource command with all subcommands wired.
func newDevicesCmd(state *appState) *cobra.Command {
	devices := &cobra.Command{
		Use:   "devices",
		Short: "Manage ActivTrak devices",
		Long:  `Commands for listing, inspecting, and managing tracked endpoint devices.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	devices.AddCommand(newDevicesListCmd(state))
	devices.AddCommand(newDevicesGetCmd(state))
	devices.AddCommand(newDevicesDeleteCmd(state))
	devices.AddCommand(newDevicesRestoreCmd(state))
	devices.AddCommand(newDevicesUninstallCmd(state))

	return devices
}

// newDevicesListCmd implements "devices list".
func newDevicesListCmd(state *appState) *cobra.Command {
	var (
		page        int
		pageSize    int
		asJSON      bool
		fieldsFlag  string
		summaryFlag bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all devices",
		RunE: func(cmd *cobra.Command, args []string) error {
			devices, err := state.client.ListDevices(cmd.Context(), page, pageSize)
			if err != nil {
				return fmt.Errorf("listing devices: %w", err)
			}

			if asJSON {
				if summaryFlag {
					return output.JSONSummary(cmd.OutOrStdout(), len(devices), nil, len(devices) == pageSize)
				}
				if fieldsFlag != "" {
					generic, err := output.ToGeneric(devices)
					if err != nil {
						return fmt.Errorf("serializing results: %w", err)
					}
					return output.JSON(cmd.OutOrStdout(), output.FilterFields(generic, strings.Split(fieldsFlag, ",")))
				}
				return output.JSON(cmd.OutOrStdout(), devices)
			}

			rows := make([][]string, len(devices))
			for i, d := range devices {
				rows[i] = []string{
					strconv.Itoa(d.ID),
					d.Hostname,
					d.Status,
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "HOSTNAME", "STATUS"}, rows)
			return nil
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 50, "Number of results per page")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	cmd.Flags().StringVar(&fieldsFlag, "fields", "", "Comma-separated top-level JSON keys to include (e.g. id,hostname)")
	cmd.Flags().BoolVar(&summaryFlag, "summary", false, "Return aggregate statistics instead of full results")

	return cmd
}

// newDevicesGetCmd implements "devices get <id>".
func newDevicesGetCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a device by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid device ID %q: must be an integer", args[0])
			}

			device, err := state.client.GetDevice(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("getting device %d: %w", id, err)
			}

			output.KeyValue(cmd.OutOrStdout(), map[string]string{
				"id":       strconv.Itoa(device.ID),
				"hostname": device.Hostname,
				"status":   device.Status,
			})
			return nil
		},
	}
}

// newDevicesDeleteCmd implements "devices delete --ids <csv>".
func newDevicesDeleteCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete devices by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.DeleteDevices(cmd.Context(), ids); err != nil {
				return fmt.Errorf("deleting devices: %w", err)
			}

			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Deleted %d devices\n", len(ids))
			} else {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "deleted")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated device IDs to delete (required)")

	return cmd
}

// newDevicesRestoreCmd implements "devices restore --ids <csv>".
func newDevicesRestoreCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore deleted devices by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.RestoreDevices(cmd.Context(), ids); err != nil {
				return fmt.Errorf("restoring devices: %w", err)
			}

			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Restored %d devices\n", len(ids))
			} else {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "restored")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated device IDs to restore (required)")

	return cmd
}

// newDevicesUninstallCmd implements "devices uninstall --ids <csv>".
func newDevicesUninstallCmd(state *appState) *cobra.Command {
	var idsFlag string

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Send uninstall request for devices by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(idsFlag)
			if err != nil {
				return fmt.Errorf("parsing --ids: %w", err)
			}
			if len(ids) == 0 {
				return fmt.Errorf("--ids is required")
			}

			if err := state.client.UninstallDevice(cmd.Context(), ids); err != nil {
				return fmt.Errorf("uninstalling devices: %w", err)
			}

			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Uninstall requested for %d devices\n", len(ids))
			} else {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "uninstall requested")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&idsFlag, "ids", "", "Comma-separated device IDs to uninstall (required)")

	return cmd
}
