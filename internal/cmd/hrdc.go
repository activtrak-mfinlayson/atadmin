package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/bulk"
	"github.com/activtrak-mfinlayson/atadmin/internal/stdin"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)

// newHRDCCmd returns the "hrdc" resource command with all subcommands wired.
func newHRDCCmd(state *appState) *cobra.Command {
	hrdc := &cobra.Command{
		Use:   "hrdc",
		Short: "Manage ActivTrak HRDC integrations",
		Long:  `Commands for interacting with the ActivTrak HR Data Connector.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	hrdc.AddCommand(newHRDCPingCmd(state))
	hrdc.AddCommand(newHRDCImportCmd(state))

	return hrdc
}

// newHRDCPingCmd implements "hrdc ping".
func newHRDCPingCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Check connectivity to the HRDC endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			if err := state.client.HRDCPing(cmd.Context()); err != nil {
				return fmt.Errorf("HRDC ping failed: %w", err)
			}

			elapsed := time.Since(start)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "OK")

			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Response time: %s\n", elapsed.Round(time.Millisecond))
			}
			return nil
		},
	}
}

// newHRDCImportCmd implements "hrdc import --file <path>".
func newHRDCImportCmd(state *appState) *cobra.Command {
	var (
		filePath  string
		fromStdin bool
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import HRDC records from a JSON or CSV file",
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

			if err := state.client.HRDCBulkImport(cmd.Context(), records); err != nil {
				return fmt.Errorf("importing HRDC records: %w", err)
			}

			if tty.IsTerminal() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Imported %d records\n", len(records))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON or CSV file")
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read JSON array from stdin instead of --file")
	return cmd
}
