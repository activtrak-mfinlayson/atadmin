package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/output"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)

// newAPIKeysCmd returns the "apikeys" resource command with all subcommands wired.
func newAPIKeysCmd(state *appState) *cobra.Command {
	apikeys := &cobra.Command{
		Use:   "apikeys",
		Short: "Manage ActivTrak API keys",
		Long:  `Commands for listing, creating, updating, and deleting ActivTrak API keys.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	apikeys.AddCommand(newAPIKeysListCmd(state))
	apikeys.AddCommand(newAPIKeysCreateCmd(state))
	apikeys.AddCommand(newAPIKeysUpdateCmd(state))
	apikeys.AddCommand(newAPIKeysDeleteCmd(state))
	apikeys.AddCommand(newAPIKeysUtilCmd(state))

	return apikeys
}

// newAPIKeysListCmd implements "apikeys list".
func newAPIKeysListCmd(state *appState) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			keys, err := state.client.ListAPIKeys(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing API keys: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), keys)
			}

			rows := make([][]string, len(keys))
			for i, k := range keys {
				rows[i] = []string{
					strconv.Itoa(k.ID),
					k.Name,
					k.KeyPrefix,
					k.LastUsedAt,
				}
			}
			output.Table(cmd.OutOrStdout(), []string{"ID", "NAME", "KEY PREFIX", "LAST USED"}, rows)
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	return cmd
}

// newAPIKeysCreateCmd implements "apikeys create --name <str>".
func newAPIKeysCreateCmd(state *appState) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			key, err := state.client.CreateAPIKey(cmd.Context(), name)
			if err != nil {
				return fmt.Errorf("creating API key: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), strconv.Itoa(key.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name for the new API key (required)")
	return cmd
}

// newAPIKeysUpdateCmd implements "apikeys update --id <id> --name <str>".
func newAPIKeysUpdateCmd(state *appState) *cobra.Command {
	var (
		id   int
		name string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an API key's name",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == 0 {
				return fmt.Errorf("--id is required")
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			if err := state.client.UpdateAPIKey(cmd.Context(), id, name); err != nil {
				return fmt.Errorf("updating API key %d: %w", id, err)
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Updated API key %d\n", id)
			return nil
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "API key ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "New name for the API key (required)")
	return cmd
}

// newAPIKeysDeleteCmd implements "apikeys delete <id>".
func newAPIKeysDeleteCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an API key by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ID %q: must be an integer", args[0])
			}

			if err := state.client.DeleteAPIKey(cmd.Context(), id); err != nil {
				return fmt.Errorf("deleting API key %d: %w", id, err)
			}

			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Deleted API key %d\n", id)
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), strconv.Itoa(id))
			}
			return nil
		},
	}
}

// newAPIKeysUtilCmd returns the "apikeys util" subcommand group.
func newAPIKeysUtilCmd(state *appState) *cobra.Command {
	util := &cobra.Command{
		Use:   "util",
		Short: "Utility operations for API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	util.AddCommand(newAPIKeysBackfillCmd(state))
	return util
}

// newAPIKeysBackfillCmd implements "apikeys util backfill [all|<id>]".
func newAPIKeysBackfillCmd(state *appState) *cobra.Command {
	backfill := &cobra.Command{
		Use:   "backfill [all|<id>]",
		Short: "Backfill instance IDs for API keys",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := args[0]
			if arg == "all" {
				if err := state.client.BackfillAllAPIKeys(cmd.Context()); err != nil {
					return fmt.Errorf("backfilling all API keys: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Backfill triggered for all API keys")
				return nil
			}

			id, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("invalid ID %q: must be an integer or 'all'", arg)
			}
			if err := state.client.BackfillAPIKey(cmd.Context(), id); err != nil {
				return fmt.Errorf("backfilling API key %d: %w", id, err)
			}
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Backfill triggered for API key %d\n", id)
			return nil
		},
	}

	return backfill
}
