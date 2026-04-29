package setup

import (
	core "dappco.re/go"
)

func TestGithubProtection_GetBranchProtection_Good(t *core.T) {
	ghHappy(t)
	protection, err := GetBranchProtection("owner/repo", "main")

	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)
}

func TestGithubProtection_GetBranchProtection_Bad(t *core.T) {
	protection, err := GetBranchProtection("invalid", "main")
	core.AssertError(t, err)

	core.AssertNil(t, protection)
	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestGithubProtection_GetBranchProtection_Ugly(t *core.T) {
	fakeGH(t, "echo '404 Branch not protected' >&2\nexit 1")
	protection, err := GetBranchProtection("owner/repo", "main")

	core.AssertNoError(t, err)
	core.AssertNil(t, protection)
}

func TestGithubProtection_SetBranchProtection_Good(t *core.T) {
	ghHappy(t)
	err := SetBranchProtection("owner/repo", "main", BranchProtectionConfig{RequiredReviews: 1, DismissStale: true})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestGithubProtection_SetBranchProtection_Bad(t *core.T) {
	err := SetBranchProtection("invalid", "main", BranchProtectionConfig{})
	core.AssertError(t, err)

	core.AssertContains(t, err.Error(), "invalid repo format")
}

func TestGithubProtection_SetBranchProtection_Ugly(t *core.T) {
	ghHappy(t)
	err := SetBranchProtection("owner/repo", "", BranchProtectionConfig{})

	core.AssertNoError(t, err)
	core.AssertTrue(t, err == nil)
}

func TestGithubProtection_SyncBranchProtection_Good(t *core.T) {
	ghHappy(t)
	cfg := &GitHubConfig{BranchProtection: []BranchProtectionConfig{{Branch: "main", RequiredReviews: 2}}}
	changes, err := SyncBranchProtection("owner/repo", cfg, true)

	core.AssertNoError(t, err)
	core.AssertEqual(t, ChangeUpdate, changes.Changes[0].Type)
}

func TestGithubProtection_SyncBranchProtection_Bad(t *core.T) {
	changes, err := SyncBranchProtection("invalid", &GitHubConfig{BranchProtection: []BranchProtectionConfig{{Branch: "main"}}}, true)
	core.AssertError(t, err)

	core.AssertNil(t, changes)
	core.AssertContains(t, err.Error(), "failed to get")
}

func TestGithubProtection_SyncBranchProtection_Ugly(t *core.T) {
	changes, err := SyncBranchProtection("owner/repo", &GitHubConfig{}, true)
	core.AssertNoError(t, err)

	core.AssertFalse(t, changes.HasChanges())
	core.AssertEmpty(t, changes.Changes)
}
