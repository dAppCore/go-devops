package setup

import (
	core "dappco.re/go"
)

func TestGithubWebhooks_ListWebhooks_Good(t *core.T) {
	ghHappy(t)
	hooks, err := ListWebhooks("owner/repo")

	core.AssertNoError(t, err)
	core.AssertEqual(t, 7, hooks[0].ID)
}

func TestGithubWebhooks_ListWebhooks_Bad(t *core.T) {
	hooks, err := ListWebhooks("invalid")
	core.AssertError(t, err)

	core.AssertNil(t, hooks)
	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestGithubWebhooks_ListWebhooks_Ugly(t *core.T) {
	fakeGH(t, "echo '[]'\nexit 0")
	hooks, err := ListWebhooks("owner/repo")

	core.AssertNoError(t, err)
	core.AssertEmpty(t, hooks)
}

func TestGithubWebhooks_CreateWebhook_Good(t *core.T) {
	ghHappy(t)
	err := CreateWebhook("owner/repo", "ci", WebhookConfig{URL: "https://hooks.example", ContentType: "json", Events: []string{"push"}})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestGithubWebhooks_CreateWebhook_Bad(t *core.T) {
	err := CreateWebhook("invalid", "ci", WebhookConfig{})
	core.AssertError(t, err)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestGithubWebhooks_CreateWebhook_Ugly(t *core.T) {
	ghHappy(t)
	active := false
	err := CreateWebhook("owner/repo", "ci", WebhookConfig{URL: "https://hooks.example", Secret: "secret", Active: &active})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestGithubWebhooks_UpdateWebhook_Good(t *core.T) {
	ghHappy(t)
	err := UpdateWebhook("owner/repo", 7, WebhookConfig{URL: "https://hooks.example", ContentType: "json", Events: []string{"push"}})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestGithubWebhooks_UpdateWebhook_Bad(t *core.T) {
	err := UpdateWebhook("invalid", 7, WebhookConfig{})
	core.AssertError(t, err)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestGithubWebhooks_UpdateWebhook_Ugly(t *core.T) {
	ghHappy(t)
	active := false
	err := UpdateWebhook("owner/repo", 0, WebhookConfig{URL: "", Active: &active})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestGithubWebhooks_SyncWebhooks_Good(t *core.T) {
	ghHappy(t)
	cfg := &GitHubConfig{Webhooks: map[string]WebhookConfig{"ci": {URL: "https://hooks.example", ContentType: "form", Events: []string{"push"}}}}
	changes, err := SyncWebhooks("owner/repo", cfg, true)

	core.AssertNoError(t, err)
	core.AssertEqual(t, ChangeUpdate, changes.Changes[0].Type)
}

func TestGithubWebhooks_SyncWebhooks_Bad(t *core.T) {
	changes, err := SyncWebhooks("invalid", &GitHubConfig{Webhooks: map[string]WebhookConfig{"ci": {URL: "x"}}}, true)
	core.AssertError(t, err)

	core.AssertNil(t, changes)
	core.AssertContains(t, err.Error(), "failed to list")
}

func TestGithubWebhooks_SyncWebhooks_Ugly(t *core.T) {
	changes, err := SyncWebhooks("owner/repo", &GitHubConfig{}, true)
	core.AssertNoError(t, err)

	core.AssertFalse(t, changes.HasChanges())
	core.AssertEmpty(t, changes.Changes)
}
