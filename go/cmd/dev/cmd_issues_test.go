package dev

import core "dappco.re/go"

func TestCmdIssues_ForgeIssue_Good(t *core.T) {
	issue := ForgeIssue{Number: 42, Title: "Document release", Labels: []string{"docs"}}

	core.AssertEqual(t, int64(42), issue.Number)
	core.AssertContains(t, issue.Labels, "docs")
}

func TestCmdIssues_ForgeIssue_Bad(t *core.T) {
	issue := ForgeIssue{}

	core.AssertEqual(t, int64(0), issue.Number)
	core.AssertEmpty(t, issue.Title)
}

func TestCmdIssues_ForgeIssue_Ugly(t *core.T) {
	issue := ForgeIssue{Assignees: []string{}}

	core.AssertEmpty(t, issue.Assignees)
	core.AssertEqual(t, "", issue.RepoName)
}
