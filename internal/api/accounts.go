package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AccountPing checks connectivity to the account API.
// GET /admin/v1/accounts/ping
func (c *Client) AccountPing(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodGet, "/admin/v1/accounts/ping", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// GetPrivacySettings returns the account privacy settings.
// GET /admin/v1/accountsettings/privacy
func (c *Client) GetPrivacySettings(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/privacy")
}

// UpdatePrivacySettings replaces the account privacy settings.
// PUT /admin/v1/accountsettings/privacy
func (c *Client) UpdatePrivacySettings(ctx context.Context, body map[string]any) error {
	return c.putSettings(ctx, "/admin/v1/accountsettings/privacy", body)
}

// GetSSOSettings returns the current SSO configuration.
// GET /admin/v1/accountsettings/sso
func (c *Client) GetSSOSettings(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/sso")
}

// UpdateSSOSettings replaces the SSO configuration.
// PUT /admin/v1/accountsettings/sso
func (c *Client) UpdateSSOSettings(ctx context.Context, body map[string]any) error {
	return c.putSettings(ctx, "/admin/v1/accountsettings/sso", body)
}

// GetSSOEnabled reports whether SSO is currently enabled for the account.
// GET /admin/v1/accountsettings/sso/enabled
func (c *Client) GetSSOEnabled(ctx context.Context) (bool, error) {
	return c.getSettingsBool(ctx, "/admin/v1/accountsettings/sso/enabled")
}

// GetSSOEligible reports whether the account is eligible to enable SSO.
// GET /admin/v1/accountsettings/sso/eligible
func (c *Client) GetSSOEligible(ctx context.Context) (bool, error) {
	return c.getSettingsBool(ctx, "/admin/v1/accountsettings/sso/eligible")
}

// GetRoleAccess returns the role-access configuration rows.
// GET /admin/v1/accountsettings/roleaccess
func (c *Client) GetRoleAccess(ctx context.Context) ([]map[string]any, error) {
	return c.getSettingsSlice(ctx, "/admin/v1/accountsettings/roleaccess")
}

// SetRoleAccess replaces the role-access configuration.
// POST /admin/v1/accountsettings/roleaccess
func (c *Client) SetRoleAccess(ctx context.Context, body []map[string]any) error {
	return c.postSettings(ctx, "/admin/v1/accountsettings/roleaccess", body)
}

// ResetRoleAccess resets role-access to the account default.
// POST /admin/v1/accountsettings/roleaccess/reset
func (c *Client) ResetRoleAccess(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/accountsettings/roleaccess/reset", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// GetRoleDateFilter returns the role-date-filter rows.
// GET /admin/v1/accountsettings/roledatefilter
func (c *Client) GetRoleDateFilter(ctx context.Context) ([]map[string]any, error) {
	return c.getSettingsSlice(ctx, "/admin/v1/accountsettings/roledatefilter")
}

// SetRoleDateFilter replaces the role-date-filter configuration.
// POST /admin/v1/accountsettings/roledatefilter
func (c *Client) SetRoleDateFilter(ctx context.Context, body []map[string]any) error {
	return c.postSettings(ctx, "/admin/v1/accountsettings/roledatefilter", body)
}

// GetTimezone returns the account timezone setting.
// GET /admin/v1/accountsettings/timezone
func (c *Client) GetTimezone(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/timezone")
}

// UpdateTimezone sets the account timezone.
// POST /admin/v1/accountsettings/timezone
func (c *Client) UpdateTimezone(ctx context.Context, body map[string]any) error {
	return c.postSettings(ctx, "/admin/v1/accountsettings/timezone", body)
}

// ListTimezones returns all available timezones.
// GET /admin/v1/accountsettings/timezones
func (c *Client) ListTimezones(ctx context.Context) ([]map[string]any, error) {
	return c.getSettingsSlice(ctx, "/admin/v1/accountsettings/timezones")
}

// GetLocalTimezone returns the show-local-timezone setting.
// GET /admin/v1/accountsettings/showlocaltimezone
func (c *Client) GetLocalTimezone(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/showlocaltimezone")
}

// UpdateLocalTimezone updates the show-local-timezone setting.
// POST /admin/v1/accountsettings/showlocaltimezone
func (c *Client) UpdateLocalTimezone(ctx context.Context, body map[string]any) error {
	return c.postSettings(ctx, "/admin/v1/accountsettings/showlocaltimezone", body)
}

// GetAgentActivityDuration returns the configured agent activity duration.
// GET /admin/v1/accountsettings/agent/activityduration
func (c *Client) GetAgentActivityDuration(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/agent/activityduration")
}

// AddAgentActivityDuration adds an agent activity duration entry for the given minutes value.
// POST /admin/v1/accountsettings/agent/activityduration/{minutes}
func (c *Client) AddAgentActivityDuration(ctx context.Context, minutes int) error {
	path := fmt.Sprintf("/admin/v1/accountsettings/agent/activityduration/%d", minutes)
	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// UpdateAgentActivityDuration updates the agent activity duration to the given minutes value.
// PUT /admin/v1/accountsettings/agent/activityduration/{minutes}
func (c *Client) UpdateAgentActivityDuration(ctx context.Context, minutes int) error {
	path := fmt.Sprintf("/admin/v1/accountsettings/agent/activityduration/%d", minutes)
	resp, err := c.doRequest(ctx, http.MethodPut, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// DeleteAgentActivityDuration removes the agent activity duration override.
// DELETE /admin/v1/accountsettings/agent/activityduration
func (c *Client) DeleteAgentActivityDuration(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/admin/v1/accountsettings/agent/activityduration", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// GetAgentAudit returns the agent audit settings.
// GET /admin/v1/accountsettings/agent/audit
func (c *Client) GetAgentAudit(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/agent/audit")
}

// UpdateAgentAudit updates the agent audit settings.
// POST /admin/v1/accountsettings/agent/audit
func (c *Client) UpdateAgentAudit(ctx context.Context, body map[string]any) error {
	return c.postSettings(ctx, "/admin/v1/accountsettings/agent/audit", body)
}

// GetPassiveTime returns the computer passive time setting.
// GET /admin/v1/accountsettings/computerpassivetime
func (c *Client) GetPassiveTime(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/computerpassivetime")
}

// UpdatePassiveTime partially updates the computer passive time setting.
// PATCH /admin/v1/accountsettings/computerpassivetime
func (c *Client) UpdatePassiveTime(ctx context.Context, body map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodPatch, "/admin/v1/accountsettings/computerpassivetime", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// BulkUpdatePassiveTime performs a bulk update of the computer passive time setting.
// POST /admin/v1/accountsettings/computerpassivetime/bulk
func (c *Client) BulkUpdatePassiveTime(ctx context.Context, body map[string]any) error {
	return c.postSettings(ctx, "/admin/v1/accountsettings/computerpassivetime/bulk", body)
}

// GetScheduleAdherence returns the schedule adherence setting.
// GET /admin/v1/accountsettings/schedule_adherence
func (c *Client) GetScheduleAdherence(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/schedule_adherence")
}

// UpdateScheduleAdherence replaces the schedule adherence setting.
// PUT /admin/v1/accountsettings/schedule_adherence
func (c *Client) UpdateScheduleAdherence(ctx context.Context, body map[string]any) error {
	return c.putSettings(ctx, "/admin/v1/accountsettings/schedule_adherence", body)
}

// GetEmailAutoDetect returns the email auto-detection setting.
// GET /admin/v1/accountsettings/emailautodetection
func (c *Client) GetEmailAutoDetect(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/emailautodetection")
}

// UpdateEmailAutoDetect updates the email auto-detection setting.
// POST /admin/v1/accountsettings/emailautodetection
func (c *Client) UpdateEmailAutoDetect(ctx context.Context, body map[string]any) error {
	return c.postSettings(ctx, "/admin/v1/accountsettings/emailautodetection", body)
}

// GetIdentityMatch returns the identity new-agent match-user setting.
// GET /admin/v1/accountsettings/identitynewagentmatchuser
func (c *Client) GetIdentityMatch(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/identitynewagentmatchuser")
}

// UpdateIdentityMatch updates the identity new-agent match-user setting.
// POST /admin/v1/accountsettings/identitynewagentmatchuser
func (c *Client) UpdateIdentityMatch(ctx context.Context, body map[string]any) error {
	return c.postSettings(ctx, "/admin/v1/accountsettings/identitynewagentmatchuser", body)
}

// GetIdentityThreshold returns the identity search active threshold days setting.
// GET /admin/v1/accountsettings/identitysearchactivethresholddays
func (c *Client) GetIdentityThreshold(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/identitysearchactivethresholddays")
}

// UpdateIdentityThreshold updates the identity search active threshold days setting.
// POST /admin/v1/accountsettings/identitysearchactivethresholddays
func (c *Client) UpdateIdentityThreshold(ctx context.Context, body map[string]any) error {
	return c.postSettings(ctx, "/admin/v1/accountsettings/identitysearchactivethresholddays", body)
}

// GetLicenseApproval returns the license approval mode setting.
// GET /admin/v1/accountsettings/licenseapprovalmode
func (c *Client) GetLicenseApproval(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/licenseapprovalmode")
}

// UpdateLicenseApproval updates the license approval mode setting.
// POST /admin/v1/accountsettings/licenseapprovalmode
func (c *Client) UpdateLicenseApproval(ctx context.Context, body map[string]any) error {
	return c.postSettings(ctx, "/admin/v1/accountsettings/licenseapprovalmode", body)
}

// GetMSPOverage returns the MSP license overage setting.
// GET /admin/v1/accountsettings/msp_license_overage
func (c *Client) GetMSPOverage(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/msp_license_overage")
}

// UpdateMSPOverage replaces the MSP license overage setting.
// PUT /admin/v1/accountsettings/msp_license_overage
func (c *Client) UpdateMSPOverage(ctx context.Context, body map[string]any) error {
	return c.putSettings(ctx, "/admin/v1/accountsettings/msp_license_overage", body)
}

// DeleteMSPOverage removes the MSP license overage override.
// DELETE /admin/v1/accountsettings/msp_license_overage
func (c *Client) DeleteMSPOverage(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/admin/v1/accountsettings/msp_license_overage", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// GetHRISSettings returns the HRIS integration settings.
// GET /admin/v1/accountsettings/hris
func (c *Client) GetHRISSettings(ctx context.Context) (map[string]any, error) {
	return c.getSettingsMap(ctx, "/admin/v1/accountsettings/hris")
}

// GetAcademyURL returns the ActivTrak Academy URL string.
// GET /admin/v1/accountsettings/academy
func (c *Client) GetAcademyURL(ctx context.Context) (string, error) {
	return c.getSettingsString(ctx, "/admin/v1/accountsettings/academy")
}

// GetAcademyWorkRampURL returns the WorkRamp-specific Academy URL string.
// GET /admin/v1/accountsettings/academy/workramp
func (c *Client) GetAcademyWorkRampURL(ctx context.Context) (string, error) {
	return c.getSettingsString(ctx, "/admin/v1/accountsettings/academy/workramp")
}

// ---------------------------------------------------------------------------
// internal helpers
// ---------------------------------------------------------------------------

// getSettingsMap issues a GET and decodes the response body as map[string]any.
func (c *Client) getSettingsMap(ctx context.Context, path string) (map[string]any, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response from %s: %w", path, err)
	}
	return result, nil
}

// getSettingsSlice issues a GET and decodes the response body as []map[string]any.
func (c *Client) getSettingsSlice(ctx context.Context, path string) ([]map[string]any, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var result []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response from %s: %w", path, err)
	}
	return result, nil
}

// getSettingsBool issues a GET and decodes the response body as a raw JSON bool.
func (c *Client) getSettingsBool(ctx context.Context, path string) (bool, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return false, err
	}
	var result bool
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decoding bool response from %s: %w", path, err)
	}
	return result, nil
}

// getSettingsString issues a GET and decodes the response body as a raw JSON string.
func (c *Client) getSettingsString(ctx context.Context, path string) (string, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return "", err
	}
	var result string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding string response from %s: %w", path, err)
	}
	return result, nil
}

// putSettings issues a PUT with the body marshalled as JSON and checks the response.
func (c *Client) putSettings(ctx context.Context, path string, body any) error {
	resp, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// postSettings issues a POST with the body marshalled as JSON and checks the response.
func (c *Client) postSettings(ctx context.Context, path string, body any) error {
	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}
