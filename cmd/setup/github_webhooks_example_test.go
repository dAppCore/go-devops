package setup

import . "dappco.re/go"

func webhooksExampleFakeGH(body string) func() {
	dir := MustCast[string](MkdirTemp("", "webhooks-gh-*"))
	WriteFile(PathJoin(dir, "gh"), []byte("#!/bin/sh\n"+body+"\n"), 0o755)
	oldPath := Getenv("PATH")
	Setenv("PATH", dir+":"+oldPath)
	return func() { Setenv("PATH", oldPath); RemoveAll(dir) }
}

func ExampleListWebhooks() {
	cleanup := webhooksExampleFakeGH("echo '[{\"id\":7,\"name\":\"web\",\"active\":true,\"events\":[\"push\"],\"config\":{\"url\":\"https://hooks.example\",\"content_type\":\"json\"}}]'")
	defer cleanup()
	hooks, err := ListWebhooks("owner/repo")
	Println(err == nil, hooks[0].ID)
	// Output: true 7
}

func ExampleCreateWebhook() {
	cleanup := webhooksExampleFakeGH("exit 0")
	defer cleanup()
	err := CreateWebhook("owner/repo", "ci", WebhookConfig{URL: "https://hooks.example", Events: []string{"push"}})
	Println(err == nil)
	// Output: true
}

func ExampleUpdateWebhook() {
	cleanup := webhooksExampleFakeGH("exit 0")
	defer cleanup()
	err := UpdateWebhook("owner/repo", 7, WebhookConfig{URL: "https://hooks.example", Events: []string{"push"}})
	Println(err == nil)
	// Output: true
}

func ExampleSyncWebhooks() {
	cleanup := webhooksExampleFakeGH("echo '[{\"id\":7,\"name\":\"web\",\"active\":true,\"events\":[\"push\"],\"config\":{\"url\":\"https://old.example\",\"content_type\":\"json\"}}]'")
	defer cleanup()
	changes, err := SyncWebhooks("owner/repo", &GitHubConfig{Webhooks: map[string]WebhookConfig{"web": {URL: "https://hooks.example", Events: []string{"push"}}}}, true)
	Println(err == nil, changes.HasChanges())
	// Output: true true
}
