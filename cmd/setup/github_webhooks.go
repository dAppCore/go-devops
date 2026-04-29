// github_webhooks.go implements GitHub webhook synchronization.
//
// Uses the gh api command for webhook operations:
//   - gh api repos/{owner}/{repo}/hooks --method GET
//   - gh api repos/{owner}/{repo}/hooks --method POST

package setup

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	log "dappco.re/go/log"
	coreexec "dappco.re/go/process/exec"
)

// GitHubWebhook represents a webhook as returned by the GitHub API.
type GitHubWebhook struct {
	ID     int                 `json:"id"`
	Name   string              `json:"name"`
	Active bool                `json:"active"`
	Events []string            `json:"events"`
	Config GitHubWebhookConfig `json:"config"`
}

// GitHubWebhookConfig contains webhook configuration details.
type GitHubWebhookConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	InsecureSSL string `json:"insecure_ssl"`
}

// ListWebhooks fetches all webhooks for a repository.
func ListWebhooks(repoFullName string) ([]GitHubWebhook, coreFailure) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil)
	}

	endpoint := core.Sprintf("repos/%s/%s/hooks", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		stderr := core.Trim(string(output))
		if core.Contains(stderr, "Must have admin rights") || core.Contains(stderr, "403") {
			return nil, cli.Err("insufficient permissions to manage webhooks (requires admin)")
		}
		return nil, cli.Err("%s", stderr)
	}

	var hooks []GitHubWebhook
	if r := core.JSONUnmarshal(output, &hooks); !r.OK {
		return nil, r.Value.(error)
	}

	return hooks, nil
}

// CreateWebhook creates a new webhook in a repository.
func CreateWebhook(repoFullName string, name string, config WebhookConfig) (_ coreFailure) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil)
	}

	// Build the webhook payload
	payload := map[string]any{
		"name":   "web",
		"active": true,
		"events": config.Events,
		"config": map[string]any{
			"url":          config.URL,
			"content_type": config.ContentType,
			"insecure_ssl": "0",
		},
	}

	if config.Active != nil {
		payload["active"] = *config.Active
	}

	if config.Secret != "" {
		configMap := payload["config"].(map[string]any)
		configMap["secret"] = config.Secret
	}

	payloadJSON := core.JSONMarshal(payload)
	if !payloadJSON.OK {
		return payloadJSON.Value.(error)
	}

	endpoint := core.Sprintf("repos/%s/%s/hooks", parts[0], parts[1])
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "POST", "--input", "-").WithStdin(core.NewReader(string(payloadJSON.Value.([]byte))))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cli.Err("%s", core.Trim(string(output)))
	}
	return nil
}

// UpdateWebhook updates an existing webhook.
func UpdateWebhook(repoFullName string, hookID int, config WebhookConfig) (_ coreFailure) {
	parts := core.Split(repoFullName, "/")
	if len(parts) != 2 {
		return log.E("setup.github", core.Sprintf("invalid repo format: %s", repoFullName), nil)
	}

	payload := map[string]any{
		"active": true,
		"events": config.Events,
		"config": map[string]any{
			"url":          config.URL,
			"content_type": config.ContentType,
			"insecure_ssl": "0",
		},
	}

	if config.Active != nil {
		payload["active"] = *config.Active
	}

	if config.Secret != "" {
		configMap := payload["config"].(map[string]any)
		configMap["secret"] = config.Secret
	}

	payloadJSON := core.JSONMarshal(payload)
	if !payloadJSON.OK {
		return payloadJSON.Value.(error)
	}

	endpoint := core.Sprintf("repos/%s/%s/hooks/%d", parts[0], parts[1], hookID)
	cmd := coreexec.Command(core.Background(), "gh", "api", endpoint, "--method", "PATCH", "--input", "-").WithStdin(core.NewReader(string(payloadJSON.Value.([]byte))))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cli.Err("%s", core.Trim(string(output)))
	}
	return nil
}

// SyncWebhooks synchronizes webhooks for a repository.
// Webhooks are matched by URL - if a webhook with the same URL exists, it's updated.
// Otherwise, a new webhook is created.
func SyncWebhooks(repoFullName string, config *GitHubConfig, dryRun bool) (*ChangeSet, coreFailure) {
	changes := NewChangeSet(repoFullName)

	// Skip if no webhooks configured
	if len(config.Webhooks) == 0 {
		return changes, nil
	}

	// Get existing webhooks
	existing, err := ListWebhooks(repoFullName)
	if err != nil {
		// If permission denied, note it but don't fail entirely
		if core.Contains(err.Error(), "insufficient permissions") {
			changes.Add(CategoryWebhook, ChangeSkip, "all", "insufficient permissions")
			return changes, nil
		}
		return nil, cli.Wrap(err, "failed to list webhooks")
	}

	// Build lookup map by URL
	existingByURL := make(map[string]GitHubWebhook)
	for _, hook := range existing {
		existingByURL[hook.Config.URL] = hook
	}

	// Process each configured webhook
	for name, wantHook := range config.Webhooks {
		// Skip webhooks with empty URLs (env var not set)
		if wantHook.URL == "" {
			changes.Add(CategoryWebhook, ChangeSkip, name, "URL not configured")
			continue
		}

		existingHook, exists := existingByURL[wantHook.URL]

		if !exists {
			// Create new webhook
			changes.Add(CategoryWebhook, ChangeCreate, name, wantHook.URL)
			if !dryRun {
				if err := CreateWebhook(repoFullName, name, wantHook); err != nil {
					return changes, cli.Wrap(err, "failed to create webhook "+name)
				}
			}
			continue
		}

		// Check if update is needed
		needsUpdate := false
		details := make(map[string]string)

		// Check events
		if !stringSliceEqual(existingHook.Events, wantHook.Events) {
			needsUpdate = true
			details["events"] = core.Sprintf("%v -> %v", existingHook.Events, wantHook.Events)
		}

		// Check content type
		if existingHook.Config.ContentType != wantHook.ContentType {
			needsUpdate = true
			details["content_type"] = core.Sprintf("%s -> %s", existingHook.Config.ContentType, wantHook.ContentType)
		}

		// Check active state
		wantActive := true
		if wantHook.Active != nil {
			wantActive = *wantHook.Active
		}
		if existingHook.Active != wantActive {
			needsUpdate = true
			details["active"] = core.Sprintf("%v -> %v", existingHook.Active, wantActive)
		}

		if needsUpdate {
			changes.AddWithDetails(CategoryWebhook, ChangeUpdate, name, "", details)
			if !dryRun {
				if err := UpdateWebhook(repoFullName, existingHook.ID, wantHook); err != nil {
					return changes, cli.Wrap(err, "failed to update webhook "+name)
				}
			}
		} else {
			changes.Add(CategoryWebhook, ChangeSkip, name, "up to date")
		}
	}

	return changes, nil
}

// stringSliceEqual compares two string slices for equality (order-independent).
// Uses frequency counting to properly handle duplicates.
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	// Count frequencies in slice a
	counts := make(map[string]int)
	for _, s := range a {
		counts[s]++
	}
	// Decrement for each element in slice b
	for _, s := range b {
		counts[s]--
		if counts[s] < 0 {
			return false
		}
	}
	// All counts should be zero if slices are equal
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}
