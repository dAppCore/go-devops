package setup

import (
	core "dappco.re/go"
)

func TestGithubSecurity_GetSecuritySettings_Good(t *core.T) {
	ghHappy(t)
	status, err := GetSecuritySettings("owner/repo")

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, status.DependabotAlerts)
}

func TestGithubSecurity_GetSecuritySettings_Bad(t *core.T) {
	status, err := GetSecuritySettings("invalid")
	core.AssertFalse(t, err.OK)

	core.AssertNil(t, status)
	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestGithubSecurity_GetSecuritySettings_Ugly(t *core.T) {
	fakeGH(t, `
if [ "$2" = "repos/owner/repo/vulnerability-alerts" ]; then echo 404 >&2; exit 1; fi
echo '{"security_and_analysis":{"secret_scanning":{"status":"enabled"},"secret_scanning_push_protection":{"status":"enabled"},"dependabot_security_updates":{"status":"enabled"}}}'
`)
	status, err := GetSecuritySettings("owner/repo")

	core.AssertTrue(t, err.OK)
	core.AssertFalse(t, status.DependabotAlerts)
	core.AssertTrue(t, status.SecretScanning)
}

func TestGithubSecurity_EnableDependabotAlerts_Good(t *core.T) {
	ghHappy(t)
	err := EnableDependabotAlerts("owner/repo")

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubSecurity_EnableDependabotAlerts_Bad(t *core.T) {
	err := EnableDependabotAlerts("invalid")
	core.AssertFalse(t, err.OK)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestGithubSecurity_EnableDependabotAlerts_Ugly(t *core.T) {
	fakeGH(t, "exit 0")
	err := EnableDependabotAlerts("owner/repo")

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubSecurity_EnableDependabotSecurityUpdates_Good(t *core.T) {
	ghHappy(t)
	err := EnableDependabotSecurityUpdates("owner/repo")

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubSecurity_EnableDependabotSecurityUpdates_Bad(t *core.T) {
	err := EnableDependabotSecurityUpdates("invalid")
	core.AssertFalse(t, err.OK)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestGithubSecurity_EnableDependabotSecurityUpdates_Ugly(t *core.T) {
	fakeGH(t, "exit 0")
	err := EnableDependabotSecurityUpdates("owner/repo")

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubSecurity_DisableDependabotSecurityUpdates_Good(t *core.T) {
	ghHappy(t)
	err := DisableDependabotSecurityUpdates("owner/repo")

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubSecurity_DisableDependabotSecurityUpdates_Bad(t *core.T) {
	err := DisableDependabotSecurityUpdates("invalid")
	core.AssertFalse(t, err.OK)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestGithubSecurity_DisableDependabotSecurityUpdates_Ugly(t *core.T) {
	fakeGH(t, "exit 0")
	err := DisableDependabotSecurityUpdates("owner/repo")

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubSecurity_UpdateSecurityAndAnalysis_Good(t *core.T) {
	ghHappy(t)
	err := UpdateSecurityAndAnalysis("owner/repo", true, true)

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubSecurity_UpdateSecurityAndAnalysis_Bad(t *core.T) {
	err := UpdateSecurityAndAnalysis("invalid", true, true)
	core.AssertFalse(t, err.OK)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestGithubSecurity_UpdateSecurityAndAnalysis_Ugly(t *core.T) {
	fakeGH(t, "echo 'secret scanning not available'\nexit 1")
	err := UpdateSecurityAndAnalysis("owner/repo", true, true)

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, err.OK)
}

func TestGithubSecurity_SyncSecuritySettings_Good(t *core.T) {
	ghHappy(t)
	cfg := &GitHubConfig{Security: SecurityConfig{DependabotSecurityUpdates: true, SecretScanning: true}}
	changes, err := SyncSecuritySettings("owner/repo", cfg, true)

	core.AssertTrue(t, err.OK)
	core.AssertTrue(t, changes.HasChanges())
}

func TestGithubSecurity_SyncSecuritySettings_Bad(t *core.T) {
	changes, err := SyncSecuritySettings("invalid", &GitHubConfig{}, true)
	core.AssertFalse(t, err.OK)

	core.AssertNil(t, changes)
	core.AssertContains(t, err.Error(), "failed to get")
}

func TestGithubSecurity_SyncSecuritySettings_Ugly(t *core.T) {
	ghHappy(t)
	changes, err := SyncSecuritySettings("owner/repo", &GitHubConfig{}, true)

	core.AssertTrue(t, err.OK)
	core.AssertFalse(t, changes.HasChanges())
}
