package setup

import . "dappco.re/go"

func protectionExampleFakeGH(body string) func() {
	dir := MustCast[string](MkdirTemp("", "protection-gh-*"))
	WriteFile(PathJoin(dir, "gh"), []byte("#!/bin/sh\n"+body+"\n"), 0o755)
	oldPath := Getenv("PATH")
	Setenv("PATH", dir+":"+oldPath)
	return func() { Setenv("PATH", oldPath); RemoveAll(dir) }
}

func ExampleGetBranchProtection() {
	cleanup := protectionExampleFakeGH("echo '{\"required_pull_request_reviews\":{\"required_approving_review_count\":1}}'")
	defer cleanup()
	protection, err := GetBranchProtection("owner/repo", "main")
	Println(err == nil, protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)
	// Output: true 1
}

func ExampleSetBranchProtection() {
	cleanup := protectionExampleFakeGH("exit 0")
	defer cleanup()
	err := SetBranchProtection("owner/repo", "main", BranchProtectionConfig{RequiredReviews: 1})
	Println(err == nil)
	// Output: true
}

func ExampleSyncBranchProtection() {
	cleanup := protectionExampleFakeGH("echo '{\"required_pull_request_reviews\":{\"required_approving_review_count\":1}}'")
	defer cleanup()
	changes, err := SyncBranchProtection("owner/repo", &GitHubConfig{BranchProtection: []BranchProtectionConfig{{Branch: "main", RequiredReviews: 2}}}, true)
	Println(err == nil, changes.HasChanges())
	// Output: true true
}
