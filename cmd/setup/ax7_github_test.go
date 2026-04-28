package setup

import core "dappco.re/go"

func ax7FakeGH(t *core.T, body string) {
	dir := t.TempDir()
	path := core.Path(dir, "gh")
	script := "#!/bin/sh\n" + body + "\n"
	core.RequireTrue(t, core.WriteFile(path, []byte(script), 0o755).OK)
	t.Setenv("PATH", dir+":"+core.Getenv("PATH"))
}

func ax7GHHappy(t *core.T) {
	ax7FakeGH(t, `
if [ "$1" = "label" ] && [ "$2" = "list" ]; then
  echo '[{"name":"bug","color":"00ff00","description":"old"}]'
  exit 0
fi
if [ "$1" = "label" ]; then
  exit 0
fi
if [ "$1" = "api" ]; then
  case "$2" in
    repos/*/*/hooks)
      if [ "$3" = "--method" ]; then exit 0; fi
      echo '[{"id":7,"name":"web","active":true,"events":["push"],"config":{"url":"https://hooks.example","content_type":"json","insecure_ssl":"0"}}]'
      exit 0
      ;;
    repos/*/*/hooks/*)
      exit 0
      ;;
    repos/*/*/branches/*/protection)
      if [ "$3" = "--method" ]; then exit 0; fi
      echo '{"required_status_checks":{"strict":true,"contexts":["test"]},"required_pull_request_reviews":{"dismiss_stale_reviews":true,"require_code_owner_reviews":true,"required_approving_review_count":1},"enforce_admins":{"enabled":true},"required_linear_history":{"enabled":false},"allow_force_pushes":{"enabled":false},"allow_deletions":{"enabled":false},"required_conversation_resolution":{"enabled":false}}'
      exit 0
      ;;
    repos/*/*/vulnerability-alerts)
      exit 0
      ;;
    repos/*/*/automated-security-fixes)
      exit 0
      ;;
    repos/*/*)
      echo '{"security_and_analysis":{"secret_scanning":{"status":"disabled"},"secret_scanning_push_protection":{"status":"disabled"},"dependabot_security_updates":{"status":"disabled"}}}'
      exit 0
      ;;
  esac
fi
echo '{}'
`)
}

func TestAX7_ListLabels_Good(t *core.T) {
	ax7GHHappy(t)
	labels, err := ListLabels("owner/repo")

	core.AssertNoError(t, err)
	core.AssertEqual(t, "bug", labels[0].Name)
}

func TestAX7_ListLabels_Bad(t *core.T) {
	ax7FakeGH(t, "echo label failure >&2\nexit 1")
	labels, err := ListLabels("owner/repo")

	core.AssertError(t, err)
	core.AssertNil(t, labels)
}

func TestAX7_ListLabels_Ugly(t *core.T) {
	ax7FakeGH(t, "echo '[]'\nexit 0")
	labels, err := ListLabels("owner/repo")

	core.AssertNoError(t, err)
	core.AssertEmpty(t, labels)
}

func TestAX7_CreateLabel_Good(t *core.T) {
	ax7GHHappy(t)
	err := CreateLabel("owner/repo", LabelConfig{Name: "bug", Color: "ff0000", Description: "Bug"})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_CreateLabel_Bad(t *core.T) {
	ax7FakeGH(t, "echo create failed\nexit 1")
	err := CreateLabel("owner/repo", LabelConfig{Name: "bug", Color: "ff0000"})

	core.AssertError(t, err)
	core.AssertContains(t, err.Error(), "create failed")
}

func TestAX7_CreateLabel_Ugly(t *core.T) {
	ax7GHHappy(t)
	err := CreateLabel("owner/repo", LabelConfig{Name: "empty-description", Color: "000000"})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_EditLabel_Good(t *core.T) {
	ax7GHHappy(t)
	err := EditLabel("owner/repo", LabelConfig{Name: "bug", Color: "ff0000", Description: "Bug"})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_EditLabel_Bad(t *core.T) {
	ax7FakeGH(t, "echo edit failed\nexit 1")
	err := EditLabel("owner/repo", LabelConfig{Name: "bug", Color: "ff0000"})

	core.AssertError(t, err)
	core.AssertContains(t, err.Error(), "edit failed")
}

func TestAX7_EditLabel_Ugly(t *core.T) {
	ax7GHHappy(t)
	err := EditLabel("owner/repo", LabelConfig{Name: "empty-description", Color: "000000"})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_SyncLabels_Good(t *core.T) {
	ax7GHHappy(t)
	cfg := &GitHubConfig{Labels: []LabelConfig{{Name: "bug", Color: "ff0000", Description: "new"}}}
	changes, err := SyncLabels("owner/repo", cfg, true)

	core.AssertNoError(t, err)
	core.AssertEqual(t, ChangeUpdate, changes.Changes[0].Type)
}

func TestAX7_SyncLabels_Bad(t *core.T) {
	ax7FakeGH(t, "echo list failed >&2\nexit 1")
	changes, err := SyncLabels("owner/repo", &GitHubConfig{}, true)

	core.AssertError(t, err)
	core.AssertNil(t, changes)
}

func TestAX7_SyncLabels_Ugly(t *core.T) {
	ax7GHHappy(t)
	changes, err := SyncLabels("owner/repo", &GitHubConfig{}, true)

	core.AssertNoError(t, err)
	core.AssertFalse(t, changes.HasChanges())
}

func TestAX7_ListWebhooks_Good(t *core.T) {
	ax7GHHappy(t)
	hooks, err := ListWebhooks("owner/repo")

	core.AssertNoError(t, err)
	core.AssertEqual(t, 7, hooks[0].ID)
}

func TestAX7_ListWebhooks_Bad(t *core.T) {
	hooks, err := ListWebhooks("invalid")
	core.AssertError(t, err)

	core.AssertNil(t, hooks)
	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestAX7_ListWebhooks_Ugly(t *core.T) {
	ax7FakeGH(t, "echo '[]'\nexit 0")
	hooks, err := ListWebhooks("owner/repo")

	core.AssertNoError(t, err)
	core.AssertEmpty(t, hooks)
}

func TestAX7_CreateWebhook_Good(t *core.T) {
	ax7GHHappy(t)
	err := CreateWebhook("owner/repo", "ci", WebhookConfig{URL: "https://hooks.example", ContentType: "json", Events: []string{"push"}})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_CreateWebhook_Bad(t *core.T) {
	err := CreateWebhook("invalid", "ci", WebhookConfig{})
	core.AssertError(t, err)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestAX7_CreateWebhook_Ugly(t *core.T) {
	ax7GHHappy(t)
	active := false
	err := CreateWebhook("owner/repo", "ci", WebhookConfig{URL: "https://hooks.example", Secret: "secret", Active: &active})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_UpdateWebhook_Good(t *core.T) {
	ax7GHHappy(t)
	err := UpdateWebhook("owner/repo", 7, WebhookConfig{URL: "https://hooks.example", ContentType: "json", Events: []string{"push"}})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_UpdateWebhook_Bad(t *core.T) {
	err := UpdateWebhook("invalid", 7, WebhookConfig{})
	core.AssertError(t, err)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestAX7_UpdateWebhook_Ugly(t *core.T) {
	ax7GHHappy(t)
	active := false
	err := UpdateWebhook("owner/repo", 0, WebhookConfig{URL: "", Active: &active})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_SyncWebhooks_Good(t *core.T) {
	ax7GHHappy(t)
	cfg := &GitHubConfig{Webhooks: map[string]WebhookConfig{"ci": {URL: "https://hooks.example", ContentType: "form", Events: []string{"push"}}}}
	changes, err := SyncWebhooks("owner/repo", cfg, true)

	core.AssertNoError(t, err)
	core.AssertEqual(t, ChangeUpdate, changes.Changes[0].Type)
}

func TestAX7_SyncWebhooks_Bad(t *core.T) {
	changes, err := SyncWebhooks("invalid", &GitHubConfig{Webhooks: map[string]WebhookConfig{"ci": {URL: "x"}}}, true)
	core.AssertError(t, err)

	core.AssertNil(t, changes)
	core.AssertContains(t, err.Error(), "failed to list")
}

func TestAX7_SyncWebhooks_Ugly(t *core.T) {
	changes, err := SyncWebhooks("owner/repo", &GitHubConfig{}, true)
	core.AssertNoError(t, err)

	core.AssertFalse(t, changes.HasChanges())
	core.AssertEmpty(t, changes.Changes)
}

func TestAX7_GetBranchProtection_Good(t *core.T) {
	ax7GHHappy(t)
	protection, err := GetBranchProtection("owner/repo", "main")

	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)
}

func TestAX7_GetBranchProtection_Bad(t *core.T) {
	protection, err := GetBranchProtection("invalid", "main")
	core.AssertError(t, err)

	core.AssertNil(t, protection)
	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestAX7_GetBranchProtection_Ugly(t *core.T) {
	ax7FakeGH(t, "echo '404 Branch not protected' >&2\nexit 1")
	protection, err := GetBranchProtection("owner/repo", "main")

	core.AssertNoError(t, err)
	core.AssertNil(t, protection)
}

func TestAX7_SetBranchProtection_Good(t *core.T) {
	ax7GHHappy(t)
	err := SetBranchProtection("owner/repo", "main", BranchProtectionConfig{RequiredReviews: 1, DismissStale: true})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_SetBranchProtection_Bad(t *core.T) {
	err := SetBranchProtection("invalid", "main", BranchProtectionConfig{})
	core.AssertError(t, err)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestAX7_SetBranchProtection_Ugly(t *core.T) {
	ax7GHHappy(t)
	err := SetBranchProtection("owner/repo", "", BranchProtectionConfig{})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_SyncBranchProtection_Good(t *core.T) {
	ax7GHHappy(t)
	cfg := &GitHubConfig{BranchProtection: []BranchProtectionConfig{{Branch: "main", RequiredReviews: 2}}}
	changes, err := SyncBranchProtection("owner/repo", cfg, true)

	core.AssertNoError(t, err)
	core.AssertEqual(t, ChangeUpdate, changes.Changes[0].Type)
}

func TestAX7_SyncBranchProtection_Bad(t *core.T) {
	changes, err := SyncBranchProtection("invalid", &GitHubConfig{BranchProtection: []BranchProtectionConfig{{Branch: "main"}}}, true)
	core.AssertError(t, err)

	core.AssertNil(t, changes)
	core.AssertContains(t, err.Error(), "failed to get")
}

func TestAX7_SyncBranchProtection_Ugly(t *core.T) {
	changes, err := SyncBranchProtection("owner/repo", &GitHubConfig{}, true)
	core.AssertNoError(t, err)

	core.AssertFalse(t, changes.HasChanges())
	core.AssertEmpty(t, changes.Changes)
}

func TestAX7_GetSecuritySettings_Good(t *core.T) {
	ax7GHHappy(t)
	status, err := GetSecuritySettings("owner/repo")

	core.AssertNoError(t, err)
	core.AssertTrue(t, status.DependabotAlerts)
}

func TestAX7_GetSecuritySettings_Bad(t *core.T) {
	status, err := GetSecuritySettings("invalid")
	core.AssertError(t, err)

	core.AssertNil(t, status)
	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestAX7_GetSecuritySettings_Ugly(t *core.T) {
	ax7FakeGH(t, `
if [ "$2" = "repos/owner/repo/vulnerability-alerts" ]; then echo 404 >&2; exit 1; fi
echo '{"security_and_analysis":{"secret_scanning":{"status":"enabled"},"secret_scanning_push_protection":{"status":"enabled"},"dependabot_security_updates":{"status":"enabled"}}}'
`)
	status, err := GetSecuritySettings("owner/repo")

	core.AssertNoError(t, err)
	core.AssertFalse(t, status.DependabotAlerts)
	core.AssertTrue(t, status.SecretScanning)
}

func TestAX7_EnableDependabotAlerts_Good(t *core.T) {
	ax7GHHappy(t)
	err := EnableDependabotAlerts("owner/repo")

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_EnableDependabotAlerts_Bad(t *core.T) {
	err := EnableDependabotAlerts("invalid")
	core.AssertError(t, err)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestAX7_EnableDependabotAlerts_Ugly(t *core.T) {
	ax7FakeGH(t, "exit 0")
	err := EnableDependabotAlerts("owner/repo")

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_EnableDependabotSecurityUpdates_Good(t *core.T) {
	ax7GHHappy(t)
	err := EnableDependabotSecurityUpdates("owner/repo")

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_EnableDependabotSecurityUpdates_Bad(t *core.T) {
	err := EnableDependabotSecurityUpdates("invalid")
	core.AssertError(t, err)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestAX7_EnableDependabotSecurityUpdates_Ugly(t *core.T) {
	ax7FakeGH(t, "exit 0")
	err := EnableDependabotSecurityUpdates("owner/repo")

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_DisableDependabotSecurityUpdates_Good(t *core.T) {
	ax7GHHappy(t)
	err := DisableDependabotSecurityUpdates("owner/repo")

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_DisableDependabotSecurityUpdates_Bad(t *core.T) {
	err := DisableDependabotSecurityUpdates("invalid")
	core.AssertError(t, err)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestAX7_DisableDependabotSecurityUpdates_Ugly(t *core.T) {
	ax7FakeGH(t, "exit 0")
	err := DisableDependabotSecurityUpdates("owner/repo")

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_UpdateSecurityAndAnalysis_Good(t *core.T) {
	ax7GHHappy(t)
	err := UpdateSecurityAndAnalysis("owner/repo", true, true)

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_UpdateSecurityAndAnalysis_Bad(t *core.T) {
	err := UpdateSecurityAndAnalysis("invalid", true, true)
	core.AssertError(t, err)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestAX7_UpdateSecurityAndAnalysis_Ugly(t *core.T) {
	ax7FakeGH(t, "echo 'secret scanning not available'\nexit 1")
	err := UpdateSecurityAndAnalysis("owner/repo", true, true)

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestAX7_SyncSecuritySettings_Good(t *core.T) {
	ax7GHHappy(t)
	cfg := &GitHubConfig{Security: SecurityConfig{DependabotSecurityUpdates: true, SecretScanning: true}}
	changes, err := SyncSecuritySettings("owner/repo", cfg, true)

	core.AssertNoError(t, err)
	core.AssertTrue(t, changes.HasChanges())
}

func TestAX7_SyncSecuritySettings_Bad(t *core.T) {
	changes, err := SyncSecuritySettings("invalid", &GitHubConfig{}, true)
	core.AssertError(t, err)

	core.AssertNil(t, changes)
	core.AssertContains(t, err.Error(), "failed to get")
}

func TestAX7_SyncSecuritySettings_Ugly(t *core.T) {
	ax7GHHappy(t)
	changes, err := SyncSecuritySettings("owner/repo", &GitHubConfig{}, true)

	core.AssertNoError(t, err)
	core.AssertFalse(t, changes.HasChanges())
}
