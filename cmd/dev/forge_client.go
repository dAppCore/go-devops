package dev

import (
	"path/filepath"
	"strings"

	"code.gitea.io/sdk/gitea"

	coreio "dappco.re/go/core/io"
	log "dappco.re/go/core/log"
	"dappco.re/go/core/scm/forge"
)

// forgeAPIClient creates a Gitea SDK client configured for the Forge instance.
// Forgejo is API-compatible with Gitea, so the Gitea SDK works directly.
func forgeAPIClient() (*gitea.Client, error) {
	forgeURL, token, err := forge.ResolveConfig("", "")
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, log.E("dev.forge", "no Forge API token configured (set FORGE_TOKEN or run: core forge config --token TOKEN)", nil)
	}
	return gitea.NewClient(forgeURL, gitea.SetToken(token))
}

// forgeRepoIdentity extracts the Forge owner/repo from a repo's git remote.
// Falls back to fallbackOrg/repoName if no forge.lthn.ai remote is found.
func forgeRepoIdentity(repoPath, fallbackOrg, repoName string) (owner, repo string) {
	configPath := filepath.Join(repoPath, ".git", "config")
	content, err := coreio.Local.Read(configPath)
	if err != nil {
		return fallbackOrg, repoName
	}

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "url = ") {
			continue
		}
		remoteURL := strings.TrimPrefix(line, "url = ")

		if !strings.Contains(remoteURL, "forge.lthn.ai") {
			continue
		}

		// ssh://git@forge.lthn.ai:2223/core/go-devops.git
		// https://forge.lthn.ai/core/go-devops.git
		parts := strings.SplitN(remoteURL, "forge.lthn.ai", 2)
		if len(parts) < 2 {
			continue
		}
		path := parts[1]

		// Remove port if present (e.g., ":2223/")
		if strings.HasPrefix(path, ":") {
			idx := strings.Index(path[1:], "/")
			if idx >= 0 {
				path = path[idx+1:]
			}
		}

		path = strings.TrimPrefix(path, "/")
		path = strings.TrimSuffix(path, ".git")

		ownerRepo := strings.SplitN(path, "/", 2)
		if len(ownerRepo) == 2 {
			return ownerRepo[0], ownerRepo[1]
		}
	}

	return fallbackOrg, repoName
}
