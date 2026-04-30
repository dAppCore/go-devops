package setup

import (
	core "dappco.re/go"
)

func fakeGH(t *core.T, body string) {
	dir := t.TempDir()
	path := core.Path(dir, "gh")
	script := "#!/bin/sh\n" + body + "\n"
	core.RequireTrue(t, core.WriteFile(path, []byte(script), 0o755).OK)
	t.Setenv("PATH", dir+":"+core.Getenv("PATH"))
}

func ghHappy(t *core.T) {
	fakeGH(t, `
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

func TestGithubLabels_ListLabels_Good(t *core.T) {
	ghHappy(t)
	labels, err := ListLabels("owner/repo")

	core.AssertTrue(t, err.OK)
	core.AssertEqual(t, "bug", labels[0].Name)
}

func TestGithubLabels_ListLabels_Bad(t *core.T) {
	fakeGH(t, "echo label failure >&2\nexit 1")
	labels, err := ListLabels("owner/repo")

	core.AssertFalse(t, err.OK)
	core.AssertNil(t, labels)
}

func TestGithubLabels_ListLabels_Ugly(t *core.T) {
	fakeGH(t, "echo '[]'\nexit 0")
	labels, err := ListLabels("owner/repo")

	core.AssertTrue(t, err.OK)
	core.AssertEmpty(t, labels)
}

func TestGithubLabels_CreateLabel_Good(t *core.T) {
	ghHappy(t)
	err := CreateLabel("owner/repo", LabelConfig{Name: "bug", Color: "ff0000", Description: "Bug"})

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubLabels_CreateLabel_Bad(t *core.T) {
	fakeGH(t, "echo create failed\nexit 1")
	err := CreateLabel("owner/repo", LabelConfig{Name: "bug", Color: "ff0000"})

	core.AssertFalse(t, err.OK)
	core.AssertContains(t, err.Error(), "create failed")
}

func TestGithubLabels_CreateLabel_Ugly(t *core.T) {
	ghHappy(t)
	err := CreateLabel("owner/repo", LabelConfig{Name: "empty-description", Color: "000000"})

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubLabels_EditLabel_Good(t *core.T) {
	ghHappy(t)
	err := EditLabel("owner/repo", LabelConfig{Name: "bug", Color: "ff0000", Description: "Bug"})

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubLabels_EditLabel_Bad(t *core.T) {
	fakeGH(t, "echo edit failed\nexit 1")
	err := EditLabel("owner/repo", LabelConfig{Name: "bug", Color: "ff0000"})

	core.AssertFalse(t, err.OK)
	core.AssertContains(t, err.Error(), "edit failed")
}

func TestGithubLabels_EditLabel_Ugly(t *core.T) {
	ghHappy(t)
	err := EditLabel("owner/repo", LabelConfig{Name: "empty-description", Color: "000000"})

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubLabels_SyncLabels_Good(t *core.T) {
	ghHappy(t)
	cfg := &GitHubConfig{Labels: []LabelConfig{{Name: "bug", Color: "ff0000", Description: "new"}}}
	changes, err := SyncLabels("owner/repo", cfg, true)

	core.AssertTrue(t, err.OK)
	core.AssertEqual(t, ChangeUpdate, changes.Changes[0].Type)
}

func TestGithubLabels_SyncLabels_Bad(t *core.T) {
	fakeGH(t, "echo list failed >&2\nexit 1")
	changes, err := SyncLabels("owner/repo", &GitHubConfig{}, true)

	core.AssertFalse(t, err.OK)
	core.AssertNil(t, changes)
}

func TestGithubLabels_SyncLabels_Ugly(t *core.T) {
	ghHappy(t)
	changes, err := SyncLabels("owner/repo", &GitHubConfig{}, true)

	core.AssertTrue(t, err.OK)
	core.AssertFalse(t, changes.HasChanges())
}
