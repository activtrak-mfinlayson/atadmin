package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/output"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)

// newSignalsCmd returns the "signals" resource command with all subcommands wired.
func newSignalsCmd(state *appState) *cobra.Command {
	signals := &cobra.Command{
		Use:   "signals",
		Short: "Manage ActivTrak signals",
		Long:  `Commands for listing, creating, updating, and deleting signals.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	signals.AddCommand(newSignalsListCmd(state))
	signals.AddCommand(newSignalsCreateCmd(state))
	signals.AddCommand(newSignalsUpdateCmd(state))
	signals.AddCommand(newSignalsDeleteCmd(state))

	return signals
}

// newSignalsListCmd implements "signals list".
func newSignalsListCmd(state *appState) *cobra.Command {
	var (
		asJSON     bool
		fieldsFlag string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all signals",
		RunE: func(cmd *cobra.Command, args []string) error {
			signals, err := state.client.ListSignals(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing signals: %w", err)
			}

			if asJSON {
				if fieldsFlag != "" {
					generic, err := output.ToGeneric(signals)
					if err != nil {
						return fmt.Errorf("serializing results: %w", err)
					}
					return output.JSON(cmd.OutOrStdout(), output.FilterFields(generic, strings.Split(fieldsFlag, ",")))
				}
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
	cmd.Flags().StringVar(&fieldsFlag, "fields", "", "Comma-separated top-level JSON keys to include (e.g. id,name)")
	return cmd
}

// newSignalsCreateCmd implements "signals create --file <path>".
func newSignalsCreateCmd(state *appState) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a signal from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			id, err := state.client.CreateSignal(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("creating signal: %w", err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), strconv.Itoa(id))
			return nil
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON file (required)")
	return cmd
}

// newSignalsUpdateCmd implements "signals update --file <path>".
func newSignalsUpdateCmd(state *appState) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a signal from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdateSignal(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating signal: %w", err)
			}
			printSuccess(cmd, "Signal updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON file (required)")
	return cmd
}

// newSignalsDeleteCmd implements "signals delete <id>".
func newSignalsDeleteCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a signal by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid signal ID %q: must be an integer", args[0])
			}
			if err := state.client.DeleteSignal(cmd.Context(), id); err != nil {
				return fmt.Errorf("deleting signal %d: %w", id, err)
			}
			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Deleted signal %d\n", id)
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), strconv.Itoa(id))
			}
			return nil
		},
	}
}
