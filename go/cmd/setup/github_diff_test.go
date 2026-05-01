package setup

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func captureStdout(t *core.T, fn func() error) (string, error) {
	buf := core.NewBuffer()
	cli.SetStdout(buf)
	defer cli.SetStdout(nil)
	t.Cleanup(func() { cli.SetStdout(nil) })

	err := fn()
	return buf.String(), err
}

func TestGithubDiff_NewChangeSet_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	core.AssertEqual(t, "repo-a", cs.Repo)

	core.AssertNotNil(t, cs.Changes)
	core.AssertEmpty(t, cs.Changes)
}

func TestGithubDiff_NewChangeSet_Bad(t *core.T) {
	cs := NewChangeSet("")
	core.AssertEqual(t, "", cs.Repo)

	core.AssertNotNil(t, cs.Changes)
	core.AssertFalse(t, cs.HasChanges())
}

func TestGithubDiff_NewChangeSet_Ugly(t *core.T) {
	first := NewChangeSet("repo-a")
	second := NewChangeSet("repo-a")
	first.Add(CategoryLabel, ChangeCreate, "bug", "create")

	core.AssertLen(t, first.Changes, 1)
	core.AssertEmpty(t, second.Changes)
}

func TestGithubDiff_ChangeSet_Add_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "create bug label")

	core.AssertLen(t, cs.Changes, 1)
	core.AssertEqual(t, ChangeCreate, cs.Changes[0].Type)
}

func TestGithubDiff_ChangeSet_Add_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeSkip, "bug", "up to date")

	core.AssertLen(t, cs.Changes, 1)
	core.AssertFalse(t, cs.HasChanges())
}

func TestGithubDiff_ChangeSet_Add_Ugly(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add("", "", "", "")

	core.AssertLen(t, cs.Changes, 1)
	core.AssertEqual(t, ChangeType(""), cs.Changes[0].Type)
}

func TestGithubDiff_ChangeSet_AddWithDetails_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.AddWithDetails(CategoryLabel, ChangeUpdate, "bug", "update", map[string]string{"color": "old -> new"})

	core.AssertLen(t, cs.Changes, 1)
	core.AssertEqual(t, "old -> new", cs.Changes[0].Details["color"])
}

func TestGithubDiff_ChangeSet_AddWithDetails_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.AddWithDetails(CategoryLabel, ChangeUpdate, "bug", "update", nil)

	core.AssertLen(t, cs.Changes, 1)
	core.AssertNil(t, cs.Changes[0].Details)
}

func TestGithubDiff_ChangeSet_AddWithDetails_Ugly(t *core.T) {
	details := map[string]string{"events": "push -> pull_request"}
	cs := NewChangeSet("repo-a")
	cs.AddWithDetails(CategoryWebhook, ChangeUpdate, "ci", "", details)

	details["events"] = "mutated"
	core.AssertEqual(t, "mutated", cs.Changes[0].Details["events"])
}

func TestGithubDiff_ChangeSet_HasChanges_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategorySecurity, ChangeCreate, "alerts", "enable")

	core.AssertTrue(t, cs.HasChanges())
	core.AssertLen(t, cs.Changes, 1)
}

func TestGithubDiff_ChangeSet_HasChanges_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategorySecurity, ChangeSkip, "alerts", "up to date")

	core.AssertFalse(t, cs.HasChanges())
	core.AssertLen(t, cs.Changes, 1)
}

func TestGithubDiff_ChangeSet_HasChanges_Ugly(t *core.T) {
	cs := NewChangeSet("repo-a")
	core.AssertFalse(t, cs.HasChanges())

	cs.Add(CategorySecurity, ChangeDelete, "alerts", "disable")
	core.AssertTrue(t, cs.HasChanges())
}

func TestGithubDiff_ChangeSet_Count_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "a", "")
	cs.Add(CategoryLabel, ChangeUpdate, "b", "")
	cs.Add(CategoryLabel, ChangeDelete, "c", "")
	cs.Add(CategoryLabel, ChangeSkip, "d", "")

	creates, updates, deletes, skips := cs.Count()
	core.AssertEqual(t, 1, creates)
	core.AssertEqual(t, 1, updates)
	core.AssertEqual(t, 1, deletes)
	core.AssertEqual(t, 1, skips)
}

func TestGithubDiff_ChangeSet_Count_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	creates, updates, deletes, skips := cs.Count()

	core.AssertEqual(t, 0, creates)
	core.AssertEqual(t, 0, updates+deletes+skips)
}

func TestGithubDiff_ChangeSet_Count_Ugly(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeType("custom"), "x", "")
	creates, updates, deletes, skips := cs.Count()

	core.AssertEqual(t, 0, creates+updates+deletes+skips)
	core.AssertLen(t, cs.Changes, 1)
}

func TestGithubDiff_ChangeSet_CountByCategory_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "")
	cs.Add(CategorySecurity, ChangeUpdate, "alerts", "")

	counts := cs.CountByCategory()
	core.AssertEqual(t, 1, counts[CategoryLabel])
	core.AssertEqual(t, 1, counts[CategorySecurity])
}

func TestGithubDiff_ChangeSet_CountByCategory_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeSkip, "bug", "")

	counts := cs.CountByCategory()
	core.AssertEmpty(t, counts)
	core.AssertEqual(t, 0, counts[CategoryLabel])
}

func TestGithubDiff_ChangeSet_CountByCategory_Ugly(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add("", ChangeCreate, "unknown", "")

	counts := cs.CountByCategory()
	core.AssertEqual(t, 1, counts[""])
	core.AssertLen(t, counts, 1)
}

func TestGithubDiff_ChangeSet_Print_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "create")
	out, err := captureStdout(t, func() error { cs.Print(true); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "repo-a")
}

func TestGithubDiff_ChangeSet_Print_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	out, err := captureStdout(t, func() error { cs.Print(false); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "repo-a")
}

func TestGithubDiff_ChangeSet_Print_Ugly(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.AddWithDetails(CategoryWebhook, ChangeUpdate, "ci", "", map[string]string{"b": "2", "a": "1"})
	out, err := captureStdout(t, func() error { cs.Print(true); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "ci")
}

func TestGithubDiff_NewAggregate_Good(t *core.T) {
	agg := NewAggregate()
	core.AssertNotNil(t, agg)

	core.AssertNotNil(t, agg.Sets)
	core.AssertEmpty(t, agg.Sets)
}

func TestGithubDiff_NewAggregate_Bad(t *core.T) {
	agg := NewAggregate()
	creates, updates, deletes, skips := agg.TotalChanges()

	core.AssertEqual(t, 0, creates)
	core.AssertEqual(t, 0, updates+deletes+skips)
}

func TestGithubDiff_NewAggregate_Ugly(t *core.T) {
	first := NewAggregate()
	second := NewAggregate()
	first.Add(NewChangeSet("repo-a"))

	core.AssertLen(t, first.Sets, 1)
	core.AssertEmpty(t, second.Sets)
}

func TestGithubDiff_Aggregate_Add_Good(t *core.T) {
	agg := NewAggregate()
	cs := NewChangeSet("repo-a")
	agg.Add(cs)

	core.AssertLen(t, agg.Sets, 1)
	core.AssertEqual(t, cs, agg.Sets[0])
}

func TestGithubDiff_Aggregate_Add_Bad(t *core.T) {
	agg := NewAggregate()
	agg.Add(nil)

	core.AssertLen(t, agg.Sets, 1)
	core.AssertNil(t, agg.Sets[0])
}

func TestGithubDiff_Aggregate_Add_Ugly(t *core.T) {
	agg := NewAggregate()
	cs := NewChangeSet("repo-a")
	agg.Add(cs)
	agg.Add(cs)

	core.AssertLen(t, agg.Sets, 2)
	core.AssertEqual(t, cs, agg.Sets[1])
}

func TestGithubDiff_Aggregate_TotalChanges_Good(t *core.T) {
	agg := NewAggregate()
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "")
	cs.Add(CategoryLabel, ChangeUpdate, "ci", "")
	agg.Add(cs)

	creates, updates, deletes, skips := agg.TotalChanges()
	core.AssertEqual(t, 1, creates)
	core.AssertEqual(t, 1, updates)
	core.AssertEqual(t, 0, deletes+skips)
}

func TestGithubDiff_Aggregate_TotalChanges_Bad(t *core.T) {
	agg := NewAggregate()
	creates, updates, deletes, skips := agg.TotalChanges()

	core.AssertEqual(t, 0, creates)
	core.AssertEqual(t, 0, updates+deletes+skips)
}

func TestGithubDiff_Aggregate_TotalChanges_Ugly(t *core.T) {
	agg := NewAggregate()
	empty := NewChangeSet("empty")
	agg.Add(empty)

	creates, updates, deletes, skips := agg.TotalChanges()
	core.AssertEqual(t, 0, creates+updates+deletes+skips)
	core.AssertLen(t, agg.Sets, 1)
}

func TestGithubDiff_Aggregate_ReposWithChanges_Good(t *core.T) {
	agg := NewAggregate()
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "")
	agg.Add(cs)

	core.AssertEqual(t, 1, agg.ReposWithChanges())
	core.AssertLen(t, agg.Sets, 1)
}

func TestGithubDiff_Aggregate_ReposWithChanges_Bad(t *core.T) {
	agg := NewAggregate()
	agg.Add(NewChangeSet("repo-a"))

	core.AssertEqual(t, 0, agg.ReposWithChanges())
	core.AssertLen(t, agg.Sets, 1)
}

func TestGithubDiff_Aggregate_ReposWithChanges_Ugly(t *core.T) {
	agg := NewAggregate()
	skipped := NewChangeSet("repo-a")
	skipped.Add(CategoryLabel, ChangeSkip, "bug", "")
	agg.Add(skipped)

	core.AssertEqual(t, 0, agg.ReposWithChanges())
	core.AssertFalse(t, skipped.HasChanges())
}

func TestGithubDiff_Aggregate_PrintSummary_Good(t *core.T) {
	agg := NewAggregate()
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "")
	agg.Add(cs)
	out, err := captureStdout(t, func() error { agg.PrintSummary(); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "1")
}

func TestGithubDiff_Aggregate_PrintSummary_Bad(t *core.T) {
	agg := NewAggregate()
	out, err := captureStdout(t, func() error { agg.PrintSummary(); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "0")
}

func TestGithubDiff_Aggregate_PrintSummary_Ugly(t *core.T) {
	agg := NewAggregate()
	skipped := NewChangeSet("repo-a")
	skipped.Add(CategoryLabel, ChangeSkip, "bug", "")
	agg.Add(skipped)
	out, err := captureStdout(t, func() error { agg.PrintSummary(); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "1")
}
