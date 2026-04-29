// github_security.go implements GitHub security settings synchronization.
//
// Uses the gh api command for security settings:
//   - gh api repos/{owner}/{repo}/vulnerability-alerts --method GET (check if enabled)
//   - gh api repos/{owner}/{repo}/vulnerability-alerts --method PUT (enable)
//   - gh api repos/{owner}/{repo}/automated-security-fixes --method PUT (enable dependabot updates)
//   - gh api repos/{owner}/{repo} --method PATCH (security_and_analysis settings)

package setup

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	log "dappco.re/go/log"
	coreexec "dappco.re/go/process/exec"
)

// GitHubSecurityStatus represents the security settings status of a repository.
type GitHubSecurityStatus struct {
	DependabotAlerts             bool
	DependabotSecurityUpdates    bool
	SecretScanning               bool
	SecretScanningPushProtection bool
}

// GitHubRepoResponse contains security-related fields from repo API.
type GitHubRepoResponse struct {
	SecurityAndAnalysis *SecurityAndAnalysis `json:"security_and_analysis"`
}

// SecurityAndAnalysis contains security feature settings.
type SecurityAndAnalysis struct {
	SecretScanning               *SecurityFeature `json:"secret_scanning"`
	SecretScanningPushProtection *SecurityFeature `json:"secret_scanning_push_protection"`
	DependabotSecurityUpdates    *SecurityFeature `json:"dependabot_security_updates"`
}

// SecurityFeature represents a single security feature status.
type SecurityFeature struct {
	Status string `json:"status"` // "enabled" or "disabled"
}

// GetSecuritySettings fetches current security settings for a repository.
func GetSecuritySettings(repoFullName string) (*GitHubSecurityStatus, coreFailure) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil)
	}

	status := &GitHubSecurityStatus{}

	// Check Dependabot alerts (vulnerability alerts)
	endpoint := core.Sprintf("repos/%s/%s/vulnerability-alerts", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "GET")
	alertsOutput, err := cmd.CombinedOutput()
	if err == nil {
		status.DependabotAlerts = true
	} else {
		stderr := string(alertsOutput)
		if core.Contains(stderr, "403") {
			return nil, cli.Err("insufficient permissions to check security settings")
		}
		status.DependabotAlerts = false
	}

	// Get repo security_and_analysis settings
	endpoint = core.Sprintf("repos/%s/%s", parts[0], parts[1])
	cmd = coreexec.Command(core.Background(), "gh", "api", endpoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, cli.Err("%s", core.Trim(string(output)))
	}

	var repo GitHubRepoResponse
	if r := core.JSONUnmarshal(output, &repo); !r.OK {
		return nil, r.Value.(error)
	}

	if repo.SecurityAndAnalysis != nil {
		if repo.SecurityAndAnalysis.SecretScanning != nil {
			status.SecretScanning = repo.SecurityAndAnalysis.SecretScanning.Status == "enabled"
		}
		if repo.SecurityAndAnalysis.SecretScanningPushProtection != nil {
			status.SecretScanningPushProtection = repo.SecurityAndAnalysis.SecretScanningPushProtection.Status == "enabled"
		}
		if repo.SecurityAndAnalysis.DependabotSecurityUpdates != nil {
			status.DependabotSecurityUpdates = repo.SecurityAndAnalysis.DependabotSecurityUpdates.Status == "enabled"
		}
	}

	return status, nil
}

// EnableDependabotAlerts enables Dependabot vulnerability alerts.
func EnableDependabotAlerts(repoFullName string) (_ coreFailure) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil)
	}

	endpoint := core.Sprintf("repos/%s/%s/vulnerability-alerts", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "PUT")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cli.Err("%s", core.Trim(string(output)))
	}
	return nil
}

// EnableDependabotSecurityUpdates enables automated Dependabot security updates.
func EnableDependabotSecurityUpdates(repoFullName string) (_ coreFailure) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil)
	}

	endpoint := core.Sprintf("repos/%s/%s/automated-security-fixes", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "PUT")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cli.Err("%s", core.Trim(string(output)))
	}
	return nil
}

// DisableDependabotSecurityUpdates disables automated Dependabot security updates.
func DisableDependabotSecurityUpdates(repoFullName string) (_ coreFailure) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil)
	}

	endpoint := core.Sprintf("repos/%s/%s/automated-security-fixes", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "DELETE")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cli.Err("%s", core.Trim(string(output)))
	}
	return nil
}

// UpdateSecurityAndAnalysis updates security_and_analysis settings.
func UpdateSecurityAndAnalysis(repoFullName string, secretScanning, pushProtection bool) (_ coreFailure) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil)
	}

	// Build the payload
	payload := map[string]any{
		"security_and_analysis": map[string]any{
			"secret_scanning": map[string]string{
				"status": boolToStatus(secretScanning),
			},
			"secret_scanning_push_protection": map[string]string{
				"status": boolToStatus(pushProtection),
			},
		},
	}

	payloadJSON := core.JSONMarshal(payload)
	if !payloadJSON.OK {
		return payloadJSON.Value.(error)
	}

	endpoint := core.Sprintf("repos/%s/%s", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "PATCH", "--input", "-")
	cmd = cmd.WithStdin(core.NewReader(string(payloadJSON.Value.([]byte))))
	output, err := cmd.CombinedOutput()
	if err != nil {
		errStr := core.Trim(string(output))
		// Some repos (private without GHAS) don't support these features
		if core.Contains(errStr, "secret scanning") || core.Contains(errStr, "not available") {
			return nil // Silently skip unsupported features
		}
		return cli.Err("%s", errStr)
	}
	return nil
}

func boolToStatus(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

// SyncSecuritySettings synchronizes security settings for a repository.
func SyncSecuritySettings(repoFullName string, config *GitHubConfig, dryRun bool) (*ChangeSet, coreFailure) {
	changes := NewChangeSet(repoFullName)

	// Get current settings
	existing, err := GetSecuritySettings(repoFullName)
	if err != nil {
		// If permission denied, note it but don't fail
		if core.Contains(err.Error(), "insufficient permissions") {
			changes.Add(CategorySecurity, ChangeSkip, "all", "insufficient permissions")
			return changes, nil
		}
		return nil, cli.Wrap(err, "failed to get security settings")
	}

	wantConfig := config.Security

	// Check Dependabot alerts
	if wantConfig.DependabotAlerts && !existing.DependabotAlerts {
		changes.Add(CategorySecurity, ChangeCreate, "dependabot_alerts", "enable")
		if !dryRun {
			if err := EnableDependabotAlerts(repoFullName); err != nil {
				return changes, cli.Wrap(err, "failed to enable dependabot alerts")
			}
		}
	} else if !wantConfig.DependabotAlerts && existing.DependabotAlerts {
		changes.Add(CategorySecurity, ChangeSkip, "dependabot_alerts", "cannot disable via API")
	} else {
		changes.Add(CategorySecurity, ChangeSkip, "dependabot_alerts", "up to date")
	}

	// Check Dependabot security updates
	if wantConfig.DependabotSecurityUpdates && !existing.DependabotSecurityUpdates {
		changes.Add(CategorySecurity, ChangeCreate, "dependabot_security_updates", "enable")
		if !dryRun {
			if err := EnableDependabotSecurityUpdates(repoFullName); err != nil {
				// This might fail if alerts aren't enabled first
				return changes, cli.Wrap(err, "failed to enable dependabot security updates")
			}
		}
	} else if !wantConfig.DependabotSecurityUpdates && existing.DependabotSecurityUpdates {
		changes.Add(CategorySecurity, ChangeDelete, "dependabot_security_updates", "disable")
		if !dryRun {
			if err := DisableDependabotSecurityUpdates(repoFullName); err != nil {
				return changes, cli.Wrap(err, "failed to disable dependabot security updates")
			}
		}
	} else {
		changes.Add(CategorySecurity, ChangeSkip, "dependabot_security_updates", "up to date")
	}

	// Check secret scanning and push protection
	needsSecurityUpdate := false
	if wantConfig.SecretScanning != existing.SecretScanning {
		needsSecurityUpdate = true
		if wantConfig.SecretScanning {
			changes.Add(CategorySecurity, ChangeCreate, "secret_scanning", "enable")
		} else {
			changes.Add(CategorySecurity, ChangeDelete, "secret_scanning", "disable")
		}
	} else {
		changes.Add(CategorySecurity, ChangeSkip, "secret_scanning", "up to date")
	}

	if wantConfig.SecretScanningPushProtection != existing.SecretScanningPushProtection {
		needsSecurityUpdate = true
		if wantConfig.SecretScanningPushProtection {
			changes.Add(CategorySecurity, ChangeCreate, "push_protection", "enable")
		} else {
			changes.Add(CategorySecurity, ChangeDelete, "push_protection", "disable")
		}
	} else {
		changes.Add(CategorySecurity, ChangeSkip, "push_protection", "up to date")
	}

	// Apply security_and_analysis changes
	if needsSecurityUpdate && !dryRun {
		if err := UpdateSecurityAndAnalysis(repoFullName, wantConfig.SecretScanning, wantConfig.SecretScanningPushProtection); err != nil {
			// Don't fail on unsupported features
			if !core.Contains(err.Error(), "not available") {
				return changes, cli.Wrap(err, "failed to update security settings")
			}
		}
	}

	return changes, nil
}
