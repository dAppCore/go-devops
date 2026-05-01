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
func GetSecuritySettings(repoFullName string) (*GitHubSecurityStatus, core.Result) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, core.Fail(log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil))
	}

	status := &GitHubSecurityStatus{}

	// Check Dependabot alerts (vulnerability alerts)
	endpoint := core.Sprintf("repos/%s/%s/vulnerability-alerts", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "GET")
	alertsOutput, alertsResult := commandCombinedOutput(cmd)
	if alertsResult.OK {
		status.DependabotAlerts = true
	} else {
		stderr := core.Trim(string(alertsOutput))
		if stderr == "" {
			stderr = alertsResult.Error()
		}
		if core.Contains(stderr, "403") {
			return nil, core.Fail(cli.Err("insufficient permissions to check security settings"))
		}
		status.DependabotAlerts = false
	}

	// Get repo security_and_analysis settings
	endpoint = core.Sprintf("repos/%s/%s", parts[0], parts[1])
	cmd = coreexec.Command(core.Background(), "gh", "api", endpoint)
	output, outputResult := commandCombinedOutput(cmd)
	if !outputResult.OK {
		return nil, core.Fail(cli.Err("%s", outputResult.Error()))
	}

	var repo GitHubRepoResponse
	if r := core.JSONUnmarshal(output, &repo); !r.OK {
		return nil, r
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

	return status, core.Ok(nil)
}

// EnableDependabotAlerts enables Dependabot vulnerability alerts.
func EnableDependabotAlerts(repoFullName string) (_ core.Result) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return core.Fail(log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil))
	}

	endpoint := core.Sprintf("repos/%s/%s/vulnerability-alerts", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "PUT")
	_, r := commandCombinedOutput(cmd)
	if !r.OK {
		return r
	}
	return core.Ok(nil)
}

// EnableDependabotSecurityUpdates enables automated Dependabot security updates.
func EnableDependabotSecurityUpdates(repoFullName string) (_ core.Result) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return core.Fail(log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil))
	}

	endpoint := core.Sprintf("repos/%s/%s/automated-security-fixes", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "PUT")
	_, r := commandCombinedOutput(cmd)
	if !r.OK {
		return r
	}
	return core.Ok(nil)
}

// DisableDependabotSecurityUpdates disables automated Dependabot security updates.
func DisableDependabotSecurityUpdates(repoFullName string) (_ core.Result) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return core.Fail(log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil))
	}

	endpoint := core.Sprintf("repos/%s/%s/automated-security-fixes", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "DELETE")
	_, r := commandCombinedOutput(cmd)
	if !r.OK {
		return r
	}
	return core.Ok(nil)
}

// UpdateSecurityAndAnalysis updates security_and_analysis settings.
func UpdateSecurityAndAnalysis(repoFullName string, secretScanning, pushProtection bool) (_ core.Result) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return core.Fail(log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil))
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
		return payloadJSON
	}

	endpoint := core.Sprintf("repos/%s/%s", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "PATCH", "--input", "-")
	cmd = cmd.WithStdin(core.NewReader(string(payloadJSON.Value.([]byte))))
	output, r := commandCombinedOutput(cmd)
	if !r.OK {
		errStr := core.Trim(string(output))
		if errStr == "" {
			errStr = r.Error()
		}
		// Some repos (private without GHAS) don't support these features
		if core.Contains(errStr, "secret scanning") || core.Contains(errStr, "not available") {
			return core.Ok(nil) // Silently skip unsupported features
		}
		return core.Fail(cli.Err("%s", errStr))
	}
	return core.Ok(nil)
}

func boolToStatus(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

// SyncSecuritySettings synchronizes security settings for a repository.
func SyncSecuritySettings(repoFullName string, config *GitHubConfig, dryRun bool) (*ChangeSet, core.Result) {
	changes := NewChangeSet(repoFullName)

	// Get current settings
	existing, r := GetSecuritySettings(repoFullName)
	if !r.OK {
		// If permission denied, note it but don't fail
		if core.Contains(r.Error(), "insufficient permissions") {
			changes.Add(CategorySecurity, ChangeSkip, "all", "insufficient permissions")
			return changes, core.Ok(nil)
		}
		return nil, core.Fail(cli.Wrap(r.Value.(error), "failed to get security settings"))
	}

	wantConfig := config.Security

	// Check Dependabot alerts
	if wantConfig.DependabotAlerts && !existing.DependabotAlerts {
		changes.Add(CategorySecurity, ChangeCreate, "dependabot_alerts", "enable")
		if !dryRun {
			if r := EnableDependabotAlerts(repoFullName); !r.OK {
				return changes, core.Fail(cli.Wrap(r.Value.(error), "failed to enable dependabot alerts"))
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
			if r := EnableDependabotSecurityUpdates(repoFullName); !r.OK {
				// This might fail if alerts aren't enabled first
				return changes, core.Fail(cli.Wrap(r.Value.(error), "failed to enable dependabot security updates"))
			}
		}
	} else if !wantConfig.DependabotSecurityUpdates && existing.DependabotSecurityUpdates {
		changes.Add(CategorySecurity, ChangeDelete, "dependabot_security_updates", "disable")
		if !dryRun {
			if r := DisableDependabotSecurityUpdates(repoFullName); !r.OK {
				return changes, core.Fail(cli.Wrap(r.Value.(error), "failed to disable dependabot security updates"))
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
		if r := UpdateSecurityAndAnalysis(repoFullName, wantConfig.SecretScanning, wantConfig.SecretScanningPushProtection); !r.OK {
			// Don't fail on unsupported features
			if !core.Contains(r.Error(), "not available") {
				return changes, core.Fail(cli.Wrap(r.Value.(error), "failed to update security settings"))
			}
		}
	}

	return changes, core.Ok(nil)
}
