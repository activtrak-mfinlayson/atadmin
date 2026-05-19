package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/activtrak-mfinlayson/atadmin/internal/output"
	"github.com/activtrak-mfinlayson/atadmin/internal/tty"
)

// newSettingsCmd returns the "settings" resource command with all subcommands wired.
func newSettingsCmd(state *appState) *cobra.Command {
	settings := &cobra.Command{
		Use:   "settings",
		Short: "Manage ActivTrak account settings",
		Long:  `Commands for reading and updating ActivTrak account-level settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	settings.AddCommand(newSettingsPingCmd(state))
	settings.AddCommand(newSettingsPrivacyCmd(state))
	settings.AddCommand(newSettingsSSOCmd(state))
	settings.AddCommand(newSettingsRoleAccessCmd(state))
	settings.AddCommand(newSettingsRoleDateFilterCmd(state))
	settings.AddCommand(newSettingsTimezoneCmd(state))
	settings.AddCommand(newSettingsTimezonesCmd(state))
	settings.AddCommand(newSettingsLocalTimezoneCmd(state))
	settings.AddCommand(newSettingsAgentDurationCmd(state))
	settings.AddCommand(newSettingsAgentAuditCmd(state))
	settings.AddCommand(newSettingsPassiveTimeCmd(state))
	settings.AddCommand(newSettingsScheduleAdherenceCmd(state))
	settings.AddCommand(newSettingsEmailAutoDetectCmd(state))
	settings.AddCommand(newSettingsIdentityMatchCmd(state))
	settings.AddCommand(newSettingsIdentityThresholdCmd(state))
	settings.AddCommand(newSettingsLicenseApprovalCmd(state))
	settings.AddCommand(newSettingsMSPOverageCmd(state))
	settings.AddCommand(newSettingsHRISCmd(state))
	settings.AddCommand(newSettingsAcademyCmd(state))

	return settings
}

// ---------------------------------------------------------------------------
// settings ping
// ---------------------------------------------------------------------------

func newSettingsPingCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Ping the account API to verify connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := state.client.AccountPing(cmd.Context()); err != nil {
				return fmt.Errorf("ping failed: %w", err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// settings privacy
// ---------------------------------------------------------------------------

func newSettingsPrivacyCmd(state *appState) *cobra.Command {
	privacy := &cobra.Command{
		Use:   "privacy",
		Short: "Manage privacy settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	privacy.AddCommand(newSettingsPrivacyGetCmd(state))
	privacy.AddCommand(newSettingsPrivacySetCmd(state))
	return privacy
}

func newSettingsPrivacyGetCmd(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get privacy settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetPrivacySettings(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting privacy settings: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	}
}

func newSettingsPrivacySetCmd(state *appState) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update privacy settings from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdatePrivacySettings(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating privacy settings: %w", err)
			}
			printSuccess(cmd, "Privacy settings updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "Path to JSON file (required)")
	return cmd
}

// ---------------------------------------------------------------------------
// settings sso
// ---------------------------------------------------------------------------

func newSettingsSSOCmd(state *appState) *cobra.Command {
	sso := &cobra.Command{
		Use:   "sso",
		Short: "Manage SSO settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	sso.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get SSO settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetSSOSettings(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting SSO settings: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update SSO settings from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdateSSOSettings(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating SSO settings: %w", err)
			}
			printSuccess(cmd, "SSO settings updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	sso.AddCommand(setCmd)

	sso.AddCommand(&cobra.Command{
		Use:   "enabled",
		Short: "Show whether SSO is enabled",
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := state.client.GetSSOEnabled(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting SSO enabled: %w", err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), strconv.FormatBool(v))
			return nil
		},
	})
	sso.AddCommand(&cobra.Command{
		Use:   "eligible",
		Short: "Show whether the account is eligible for SSO",
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := state.client.GetSSOEligible(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting SSO eligible: %w", err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), strconv.FormatBool(v))
			return nil
		},
	})
	return sso
}

// ---------------------------------------------------------------------------
// settings role-access
// ---------------------------------------------------------------------------

func newSettingsRoleAccessCmd(state *appState) *cobra.Command {
	ra := &cobra.Command{
		Use:   "role-access",
		Short: "Manage role-access settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	ra.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get role-access settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			rows, err := state.client.GetRoleAccess(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting role-access settings: %w", err)
			}
			tableRows := make([][]string, 0, len(rows))
			for _, r := range rows {
				resource := fmt.Sprintf("%v", r["resource"])
				roles := fmt.Sprintf("%v", r["roles"])
				tableRows = append(tableRows, []string{resource, roles})
			}
			output.Table(cmd.OutOrStdout(), []string{"RESOURCE", "ROLES"}, tableRows)
			return nil
		},
	})

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update role-access settings from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONArrayFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.SetRoleAccess(cmd.Context(), body); err != nil {
				return fmt.Errorf("setting role-access: %w", err)
			}
			printSuccess(cmd, "Role-access settings updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	ra.AddCommand(setCmd)

	ra.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Reset role-access to account defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := state.client.ResetRoleAccess(cmd.Context()); err != nil {
				return fmt.Errorf("resetting role-access: %w", err)
			}
			printSuccess(cmd, "Role-access reset to defaults")
			return nil
		},
	})
	return ra
}

// ---------------------------------------------------------------------------
// settings role-date-filter
// ---------------------------------------------------------------------------

func newSettingsRoleDateFilterCmd(state *appState) *cobra.Command {
	rdf := &cobra.Command{
		Use:   "role-date-filter",
		Short: "Manage role date-filter settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	rdf.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get role date-filter settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			rows, err := state.client.GetRoleDateFilter(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting role date-filter: %w", err)
			}
			return output.JSON(cmd.OutOrStdout(), rows)
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update role date-filter settings from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONArrayFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.SetRoleDateFilter(cmd.Context(), body); err != nil {
				return fmt.Errorf("setting role date-filter: %w", err)
			}
			printSuccess(cmd, "Role date-filter updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	rdf.AddCommand(setCmd)
	return rdf
}

// ---------------------------------------------------------------------------
// settings timezone
// ---------------------------------------------------------------------------

func newSettingsTimezoneCmd(state *appState) *cobra.Command {
	tz := &cobra.Command{
		Use:   "timezone",
		Short: "Manage account timezone",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	tz.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get the account timezone",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetTimezone(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting timezone: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update the account timezone from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdateTimezone(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating timezone: %w", err)
			}
			printSuccess(cmd, "Timezone updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	tz.AddCommand(setCmd)

	return tz
}

func newSettingsTimezonesCmd(state *appState) *cobra.Command {
	tzs := &cobra.Command{
		Use:   "timezones",
		Short: "List available timezones",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	tzs.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all available timezone identifiers",
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := state.client.ListTimezones(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing timezones: %w", err)
			}
			rows := make([][]string, 0, len(items))
			for _, t := range items {
				id := fmt.Sprintf("%v", t["id"])
				rows = append(rows, []string{id})
			}
			output.Table(cmd.OutOrStdout(), []string{"TIMEZONE"}, rows)
			return nil
		},
	})
	return tzs
}

// ---------------------------------------------------------------------------
// settings local-timezone
// ---------------------------------------------------------------------------

func newSettingsLocalTimezoneCmd(state *appState) *cobra.Command {
	ltz := &cobra.Command{
		Use:   "local-timezone",
		Short: "Manage local timezone display setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	ltz.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get the local-timezone setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetLocalTimezone(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting local-timezone setting: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update the local-timezone setting from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdateLocalTimezone(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating local-timezone setting: %w", err)
			}
			printSuccess(cmd, "Local-timezone setting updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	ltz.AddCommand(setCmd)
	return ltz
}

// ---------------------------------------------------------------------------
// settings agent-duration
// ---------------------------------------------------------------------------

func newSettingsAgentDurationCmd(state *appState) *cobra.Command {
	ad := &cobra.Command{
		Use:   "agent-duration",
		Short: "Manage agent activity duration setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	ad.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get the agent activity duration setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetAgentActivityDuration(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting agent activity duration: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set the agent activity duration in minutes",
		RunE: func(cmd *cobra.Command, args []string) error {
			minutes, _ := cmd.Flags().GetInt("minutes")
			if minutes <= 0 {
				return fmt.Errorf("--minutes must be a positive integer")
			}
			// Try update first; fall back to add is caller responsibility.
			if err := state.client.UpdateAgentActivityDuration(cmd.Context(), minutes); err != nil {
				return fmt.Errorf("setting agent activity duration: %w", err)
			}
			printSuccess(cmd, fmt.Sprintf("Agent activity duration set to %d minutes", minutes))
			return nil
		},
	}
	setCmd.Flags().Int("minutes", 0, "Duration in minutes (required)")
	ad.AddCommand(setCmd)

	ad.AddCommand(&cobra.Command{
		Use:   "delete",
		Short: "Delete the agent activity duration override",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := state.client.DeleteAgentActivityDuration(cmd.Context()); err != nil {
				return fmt.Errorf("deleting agent activity duration: %w", err)
			}
			printSuccess(cmd, "Agent activity duration override deleted")
			return nil
		},
	})
	return ad
}

// ---------------------------------------------------------------------------
// settings agent-audit
// ---------------------------------------------------------------------------

func newSettingsAgentAuditCmd(state *appState) *cobra.Command {
	aa := &cobra.Command{
		Use:   "agent-audit",
		Short: "Manage agent audit settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	aa.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get agent audit settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetAgentAudit(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting agent audit settings: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update agent audit settings from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdateAgentAudit(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating agent audit settings: %w", err)
			}
			printSuccess(cmd, "Agent audit settings updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	aa.AddCommand(setCmd)
	return aa
}

// ---------------------------------------------------------------------------
// settings passive-time
// ---------------------------------------------------------------------------

func newSettingsPassiveTimeCmd(state *appState) *cobra.Command {
	pt := &cobra.Command{
		Use:   "passive-time",
		Short: "Manage computer passive time settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	pt.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get the computer passive time setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetPassiveTime(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting passive time: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update passive time setting from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdatePassiveTime(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating passive time: %w", err)
			}
			printSuccess(cmd, "Passive time setting updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	pt.AddCommand(setCmd)

	bulkSetCmd := &cobra.Command{
		Use:   "bulk-set",
		Short: "Bulk-update passive time settings from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.BulkUpdatePassiveTime(cmd.Context(), body); err != nil {
				return fmt.Errorf("bulk-updating passive time: %w", err)
			}
			printSuccess(cmd, "Passive time bulk-updated")
			return nil
		},
	}
	bulkSetCmd.Flags().String("file", "", "Path to JSON file (required)")
	pt.AddCommand(bulkSetCmd)
	return pt
}

// ---------------------------------------------------------------------------
// settings schedule-adherence
// ---------------------------------------------------------------------------

func newSettingsScheduleAdherenceCmd(state *appState) *cobra.Command {
	sa := &cobra.Command{
		Use:   "schedule-adherence",
		Short: "Manage schedule adherence setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	sa.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get schedule adherence setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetScheduleAdherence(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting schedule adherence: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update schedule adherence setting from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdateScheduleAdherence(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating schedule adherence: %w", err)
			}
			printSuccess(cmd, "Schedule adherence updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	sa.AddCommand(setCmd)
	return sa
}

// ---------------------------------------------------------------------------
// settings email-autodetect
// ---------------------------------------------------------------------------

func newSettingsEmailAutoDetectCmd(state *appState) *cobra.Command {
	ea := &cobra.Command{
		Use:   "email-autodetect",
		Short: "Manage email auto-detection setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	ea.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get email auto-detection setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetEmailAutoDetect(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting email auto-detection: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update email auto-detection setting from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdateEmailAutoDetect(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating email auto-detection: %w", err)
			}
			printSuccess(cmd, "Email auto-detection setting updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	ea.AddCommand(setCmd)
	return ea
}

// ---------------------------------------------------------------------------
// settings identity-match
// ---------------------------------------------------------------------------

func newSettingsIdentityMatchCmd(state *appState) *cobra.Command {
	im := &cobra.Command{
		Use:   "identity-match",
		Short: "Manage identity new-agent match-user setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	im.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get identity match setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetIdentityMatch(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting identity match setting: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update identity match setting from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdateIdentityMatch(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating identity match: %w", err)
			}
			printSuccess(cmd, "Identity match setting updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	im.AddCommand(setCmd)
	return im
}

// ---------------------------------------------------------------------------
// settings identity-threshold
// ---------------------------------------------------------------------------

func newSettingsIdentityThresholdCmd(state *appState) *cobra.Command {
	it := &cobra.Command{
		Use:   "identity-threshold",
		Short: "Manage identity search active threshold setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	it.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get identity threshold setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetIdentityThreshold(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting identity threshold setting: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update identity threshold setting from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdateIdentityThreshold(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating identity threshold: %w", err)
			}
			printSuccess(cmd, "Identity threshold setting updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	it.AddCommand(setCmd)
	return it
}

// ---------------------------------------------------------------------------
// settings license-approval
// ---------------------------------------------------------------------------

func newSettingsLicenseApprovalCmd(state *appState) *cobra.Command {
	la := &cobra.Command{
		Use:   "license-approval",
		Short: "Manage license approval mode setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	la.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get license approval mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetLicenseApproval(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting license approval mode: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update license approval mode from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdateLicenseApproval(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating license approval mode: %w", err)
			}
			printSuccess(cmd, "License approval mode updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	la.AddCommand(setCmd)
	return la
}

// ---------------------------------------------------------------------------
// settings msp-overage
// ---------------------------------------------------------------------------

func newSettingsMSPOverageCmd(state *appState) *cobra.Command {
	mo := &cobra.Command{
		Use:   "msp-overage",
		Short: "Manage MSP license overage setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	mo.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get MSP overage setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetMSPOverage(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting MSP overage: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update MSP overage setting from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}
			body, err := readJSONObjectFile(filePath)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", filePath, err)
			}
			if err := state.client.UpdateMSPOverage(cmd.Context(), body); err != nil {
				return fmt.Errorf("updating MSP overage: %w", err)
			}
			printSuccess(cmd, "MSP overage setting updated")
			return nil
		},
	}
	setCmd.Flags().String("file", "", "Path to JSON file (required)")
	mo.AddCommand(setCmd)

	mo.AddCommand(&cobra.Command{
		Use:   "delete",
		Short: "Delete the MSP overage override",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := state.client.DeleteMSPOverage(cmd.Context()); err != nil {
				return fmt.Errorf("deleting MSP overage: %w", err)
			}
			printSuccess(cmd, "MSP overage override deleted")
			return nil
		},
	})
	return mo
}

// ---------------------------------------------------------------------------
// settings hris
// ---------------------------------------------------------------------------

func newSettingsHRISCmd(state *appState) *cobra.Command {
	hris := &cobra.Command{
		Use:   "hris",
		Short: "Manage HRIS integration settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	hris.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get HRIS integration settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := state.client.GetHRISSettings(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting HRIS settings: %w", err)
			}
			output.KeyValue(cmd.OutOrStdout(), mapAnyToString(m))
			return nil
		},
	})
	return hris
}

// ---------------------------------------------------------------------------
// settings academy
// ---------------------------------------------------------------------------

func newSettingsAcademyCmd(state *appState) *cobra.Command {
	academy := &cobra.Command{
		Use:   "academy",
		Short: "Manage ActivTrak Academy URL settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	academy.AddCommand(&cobra.Command{
		Use:   "url",
		Short: "Get the ActivTrak Academy URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := state.client.GetAcademyURL(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting Academy URL: %w", err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), u)
			return nil
		},
	})
	academy.AddCommand(&cobra.Command{
		Use:   "workramp-url",
		Short: "Get the WorkRamp Academy URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := state.client.GetAcademyWorkRampURL(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting WorkRamp Academy URL: %w", err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), u)
			return nil
		},
	})
	return academy
}

// ---------------------------------------------------------------------------
// shared helpers
// ---------------------------------------------------------------------------

// readJSONObjectFile reads a file and unmarshals it as map[string]any.
func readJSONObjectFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return m, nil
}

// readJSONArrayFile reads a file and unmarshals it as []map[string]any.
func readJSONArrayFile(path string) ([]map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	var arr []map[string]any
	if err := json.Unmarshal(data, &arr); err != nil {
		return nil, fmt.Errorf("parsing JSON array: %w", err)
	}
	return arr, nil
}

// mapAnyToString converts map[string]any values to map[string]string for KeyValue output.
func mapAnyToString(m map[string]any) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// printSuccess writes a success message to stderr in TTY mode, or nothing in script mode.
func printSuccess(cmd *cobra.Command, msg string) {
	if tty.IsTerminal() {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), msg)
	}
}
