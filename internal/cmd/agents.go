package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/api"
	"github.com/activtrak-mfinlayson/atadmin/internal/output"
)

// newAgentsCmd returns the "agents" resource command with all subcommands wired.
func newAgentsCmd(state *appState) *cobra.Command {
	agents := &cobra.Command{
		Use:   "agents",
		Short: "List and inspect agent (device) entities",
		Long:  `Commands for listing and inspecting agent entities registered in the account.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	agents.AddCommand(newAgentsListCmd(state))
	return agents
}

// newAgentsListCmd implements "agents list".
func newAgentsListCmd(state *appState) *cobra.Command {
	var (
		filter     string
		search     string
		searchType string
		sort       string
		sortDir    string
		limit      int
		cursor     string
		asJSON     bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List agent (device) entities",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			page, err := state.client.ListAgents(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("listing agents: %w", err)
			}

			if asJSON {
				return output.JSON(cmd.OutOrStdout(), page)
			}

			rows := make([][]string, len(page.Results))
			for i, u := range page.Results {
				// Display the first associated agent's device details when available.
				var userID, userName, domain, alias, license, lastLog string
				userID = strconv.FormatInt(u.ID, 10)
				if len(u.Agents) > 0 {
					a := u.Agents[0]
					userName = a.UserName
					domain = a.LogonDomain
					alias = a.Alias
					license = a.LicenseStatus
					lastLog = shortDatetime(a.LastLog)
				}
				rows[i] = []string{userID, userName, domain, alias, license, lastLog}
			}
			output.Table(cmd.OutOrStdout(), []string{"USER ID", "USERNAME", "DOMAIN", "ALIAS", "LICENSE", "LAST LOG"}, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter type (tracked, untracked, usersWithAgents, ...)")
	cmd.Flags().StringVar(&search, "search", "", "Search term")
	cmd.Flags().StringVar(&searchType, "search-type", "", "Field to search")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort field")
	cmd.Flags().StringVar(&sortDir, "sort-dir", "", "Sort direction: asc or desc")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum results (0 = server default)")
	cmd.Flags().StringVar(&cursor, "cursor", "", "Pagination cursor")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")

	return cmd
}

// shortDatetime trims an ISO datetime to the date portion for compact display.
func shortDatetime(s string) string {
	if s == "" {
		return ""
	}
	if idx := strings.IndexByte(s, 'T'); idx > 0 {
		return s[:idx]
	}
	return s
}
