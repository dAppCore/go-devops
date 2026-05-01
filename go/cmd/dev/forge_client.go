package dev

import (
	"code.gitea.io/sdk/gitea"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
	log "dappco.re/go/log"
	"dappco.re/go/scm/forge"
)

// forgeAPIClient creates a Gitea SDK client configured for the Forge instance.
// Forgejo is API-compatible with Gitea, so the Gitea SDK works directly.
func forgeAPIClient() (*gitea.Client, core.Result) {
	forgeURL, token, err := forge.ResolveConfig("", "")
	if err != nil {
		return nil, core.Fail(err)
	}
	if token == "" {
		return nil, core.Fail(log.E("dev.forge", "no Forge API token configured (set FORGE_TOKEN or run: core forge config --token TOKEN)", nil))
	}
	client, err := gitea.NewClient(forgeURL, gitea.SetToken(token))
	if err != nil {
		return nil, core.Fail(err)
	}
	return client, core.Ok(nil)
}

// forgeRepoIdentity extracts the Forge owner/repo from a repo's git remote.
// Falls back to fallbackOrg/repoName if no forge.lthn.ai remote is found.
func forgeRepoIdentity(repoPath, fallbackOrg, repoName string) (owner, repo string) {
	configPath := core.PathJoin(repoPath, ".git", "config")
	content, err := coreio.Local.Read(configPath)
	if err != nil {
		return fallbackOrg, repoName
	}

	for _, line := range core.Split(content, "\n") {
		line = core.Trim(line)
		if !core.HasPrefix(line, "url = ") {
			continue
		}
		remoteURL := core.TrimPrefix(line, "url = ")

		if !core.Contains(remoteURL, "forge.lthn.ai") {
			continue
		}

		// ssh://git@forge.lthn.ai:2223/core/go-devops.git
		// https://forge.lthn.ai/core/go-devops.git
		parts := core.SplitN(remoteURL, "forge.lthn.ai", 2)
		if len(parts) < 2 {
			continue
		}
		path := parts[1]

		// Remove port if present (e.g., ":2223/")
		if core.HasPrefix(path, ":") {
			portParts := core.SplitN(path[1:], "/", 2)
			if len(portParts) == 2 {
				path = portParts[1]
			}
		}

		path = core.TrimPrefix(path, "/")
		path = core.TrimSuffix(path, ".git")

		ownerRepo := core.SplitN(path, "/", 2)
		if len(ownerRepo) == 2 {
			return ownerRepo[0], ownerRepo[1]
		}
	}

	return fallbackOrg, repoName
}
