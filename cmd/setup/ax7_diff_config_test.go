package setup

import core "dappco.re/go"

func TestAX7_NewChangeSet_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	core.AssertEqual(t, "repo-a", cs.Repo)

	core.AssertNotNil(t, cs.Changes)
	core.AssertEmpty(t, cs.Changes)
}

func TestAX7_NewChangeSet_Bad(t *core.T) {
	cs := NewChangeSet("")
	core.AssertEqual(t, "", cs.Repo)

	core.AssertNotNil(t, cs.Changes)
	core.AssertFalse(t, cs.HasChanges())
}

func TestAX7_NewChangeSet_Ugly(t *core.T) {
	first := NewChangeSet("repo-a")
	second := NewChangeSet("repo-a")
	first.Add(CategoryLabel, ChangeCreate, "bug", "create")

	core.AssertLen(t, first.Changes, 1)
	core.AssertEmpty(t, second.Changes)
}

func TestAX7_ChangeSet_Add_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "create bug label")

	core.AssertLen(t, cs.Changes, 1)
	core.AssertEqual(t, ChangeCreate, cs.Changes[0].Type)
}

func TestAX7_ChangeSet_Add_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeSkip, "bug", "up to date")

	core.AssertLen(t, cs.Changes, 1)
	core.AssertFalse(t, cs.HasChanges())
}

func TestAX7_ChangeSet_Add_Ugly(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add("", "", "", "")

	core.AssertLen(t, cs.Changes, 1)
	core.AssertEqual(t, ChangeType(""), cs.Changes[0].Type)
}

func TestAX7_ChangeSet_AddWithDetails_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.AddWithDetails(CategoryLabel, ChangeUpdate, "bug", "update", map[string]string{"color": "old -> new"})

	core.AssertLen(t, cs.Changes, 1)
	core.AssertEqual(t, "old -> new", cs.Changes[0].Details["color"])
}

func TestAX7_ChangeSet_AddWithDetails_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.AddWithDetails(CategoryLabel, ChangeUpdate, "bug", "update", nil)

	core.AssertLen(t, cs.Changes, 1)
	core.AssertNil(t, cs.Changes[0].Details)
}

func TestAX7_ChangeSet_AddWithDetails_Ugly(t *core.T) {
	details := map[string]string{"events": "push -> pull_request"}
	cs := NewChangeSet("repo-a")
	cs.AddWithDetails(CategoryWebhook, ChangeUpdate, "ci", "", details)

	details["events"] = "mutated"
	core.AssertEqual(t, "mutated", cs.Changes[0].Details["events"])
}

func TestAX7_ChangeSet_HasChanges_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategorySecurity, ChangeCreate, "alerts", "enable")

	core.AssertTrue(t, cs.HasChanges())
	core.AssertLen(t, cs.Changes, 1)
}

func TestAX7_ChangeSet_HasChanges_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategorySecurity, ChangeSkip, "alerts", "up to date")

	core.AssertFalse(t, cs.HasChanges())
	core.AssertLen(t, cs.Changes, 1)
}

func TestAX7_ChangeSet_HasChanges_Ugly(t *core.T) {
	cs := NewChangeSet("repo-a")
	core.AssertFalse(t, cs.HasChanges())

	cs.Add(CategorySecurity, ChangeDelete, "alerts", "disable")
	core.AssertTrue(t, cs.HasChanges())
}

func TestAX7_ChangeSet_Count_Good(t *core.T) {
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

func TestAX7_ChangeSet_Count_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	creates, updates, deletes, skips := cs.Count()

	core.AssertEqual(t, 0, creates)
	core.AssertEqual(t, 0, updates+deletes+skips)
}

func TestAX7_ChangeSet_Count_Ugly(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeType("custom"), "x", "")
	creates, updates, deletes, skips := cs.Count()

	core.AssertEqual(t, 0, creates+updates+deletes+skips)
	core.AssertLen(t, cs.Changes, 1)
}

func TestAX7_ChangeSet_CountByCategory_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "")
	cs.Add(CategorySecurity, ChangeUpdate, "alerts", "")

	counts := cs.CountByCategory()
	core.AssertEqual(t, 1, counts[CategoryLabel])
	core.AssertEqual(t, 1, counts[CategorySecurity])
}

func TestAX7_ChangeSet_CountByCategory_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeSkip, "bug", "")

	counts := cs.CountByCategory()
	core.AssertEmpty(t, counts)
	core.AssertEqual(t, 0, counts[CategoryLabel])
}

func TestAX7_ChangeSet_CountByCategory_Ugly(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add("", ChangeCreate, "unknown", "")

	counts := cs.CountByCategory()
	core.AssertEqual(t, 1, counts[""])
	core.AssertLen(t, counts, 1)
}

func TestAX7_ChangeSet_Print_Good(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "create")
	out, err := captureStdout(t, func() error { cs.Print(true); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "repo-a")
}

func TestAX7_ChangeSet_Print_Bad(t *core.T) {
	cs := NewChangeSet("repo-a")
	out, err := captureStdout(t, func() error { cs.Print(false); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "repo-a")
}

func TestAX7_ChangeSet_Print_Ugly(t *core.T) {
	cs := NewChangeSet("repo-a")
	cs.AddWithDetails(CategoryWebhook, ChangeUpdate, "ci", "", map[string]string{"b": "2", "a": "1"})
	out, err := captureStdout(t, func() error { cs.Print(true); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "ci")
}

func TestAX7_NewAggregate_Good(t *core.T) {
	agg := NewAggregate()
	core.AssertNotNil(t, agg)

	core.AssertNotNil(t, agg.Sets)
	core.AssertEmpty(t, agg.Sets)
}

func TestAX7_NewAggregate_Bad(t *core.T) {
	agg := NewAggregate()
	creates, updates, deletes, skips := agg.TotalChanges()

	core.AssertEqual(t, 0, creates)
	core.AssertEqual(t, 0, updates+deletes+skips)
}

func TestAX7_NewAggregate_Ugly(t *core.T) {
	first := NewAggregate()
	second := NewAggregate()
	first.Add(NewChangeSet("repo-a"))

	core.AssertLen(t, first.Sets, 1)
	core.AssertEmpty(t, second.Sets)
}

func TestAX7_Aggregate_Add_Good(t *core.T) {
	agg := NewAggregate()
	cs := NewChangeSet("repo-a")
	agg.Add(cs)

	core.AssertLen(t, agg.Sets, 1)
	core.AssertEqual(t, cs, agg.Sets[0])
}

func TestAX7_Aggregate_Add_Bad(t *core.T) {
	agg := NewAggregate()
	agg.Add(nil)

	core.AssertLen(t, agg.Sets, 1)
	core.AssertNil(t, agg.Sets[0])
}

func TestAX7_Aggregate_Add_Ugly(t *core.T) {
	agg := NewAggregate()
	cs := NewChangeSet("repo-a")
	agg.Add(cs)
	agg.Add(cs)

	core.AssertLen(t, agg.Sets, 2)
	core.AssertEqual(t, cs, agg.Sets[1])
}

func TestAX7_Aggregate_TotalChanges_Good(t *core.T) {
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

func TestAX7_Aggregate_TotalChanges_Bad(t *core.T) {
	agg := NewAggregate()
	creates, updates, deletes, skips := agg.TotalChanges()

	core.AssertEqual(t, 0, creates)
	core.AssertEqual(t, 0, updates+deletes+skips)
}

func TestAX7_Aggregate_TotalChanges_Ugly(t *core.T) {
	agg := NewAggregate()
	empty := NewChangeSet("empty")
	agg.Add(empty)

	creates, updates, deletes, skips := agg.TotalChanges()
	core.AssertEqual(t, 0, creates+updates+deletes+skips)
	core.AssertLen(t, agg.Sets, 1)
}

func TestAX7_Aggregate_ReposWithChanges_Good(t *core.T) {
	agg := NewAggregate()
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "")
	agg.Add(cs)

	core.AssertEqual(t, 1, agg.ReposWithChanges())
	core.AssertLen(t, agg.Sets, 1)
}

func TestAX7_Aggregate_ReposWithChanges_Bad(t *core.T) {
	agg := NewAggregate()
	agg.Add(NewChangeSet("repo-a"))

	core.AssertEqual(t, 0, agg.ReposWithChanges())
	core.AssertLen(t, agg.Sets, 1)
}

func TestAX7_Aggregate_ReposWithChanges_Ugly(t *core.T) {
	agg := NewAggregate()
	skipped := NewChangeSet("repo-a")
	skipped.Add(CategoryLabel, ChangeSkip, "bug", "")
	agg.Add(skipped)

	core.AssertEqual(t, 0, agg.ReposWithChanges())
	core.AssertFalse(t, skipped.HasChanges())
}

func TestAX7_Aggregate_PrintSummary_Good(t *core.T) {
	agg := NewAggregate()
	cs := NewChangeSet("repo-a")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "")
	agg.Add(cs)
	out, err := captureStdout(t, func() error { agg.PrintSummary(); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "1")
}

func TestAX7_Aggregate_PrintSummary_Bad(t *core.T) {
	agg := NewAggregate()
	out, err := captureStdout(t, func() error { agg.PrintSummary(); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "0")
}

func TestAX7_Aggregate_PrintSummary_Ugly(t *core.T) {
	agg := NewAggregate()
	skipped := NewChangeSet("repo-a")
	skipped.Add(CategoryLabel, ChangeSkip, "bug", "")
	agg.Add(skipped)
	out, err := captureStdout(t, func() error { agg.PrintSummary(); return nil })

	core.AssertNoError(t, err)
	core.AssertContains(t, out, "1")
}

func TestAX7_DefaultCIConfig_Good(t *core.T) {
	cfg := DefaultCIConfig()
	core.AssertEqual(t, "host-uk/tap", cfg.Tap)

	core.AssertEqual(t, "core", cfg.Formula)
	core.AssertEqual(t, "dev", cfg.DefaultVersion)
}

func TestAX7_DefaultCIConfig_Bad(t *core.T) {
	cfg := DefaultCIConfig()
	cfg.DefaultVersion = ""

	core.AssertEqual(t, "", cfg.DefaultVersion)
	core.AssertEqual(t, "core-cli", cfg.ChocolateyPkg)
}

func TestAX7_DefaultCIConfig_Ugly(t *core.T) {
	first := DefaultCIConfig()
	second := DefaultCIConfig()
	first.Formula = "mutated"

	core.AssertEqual(t, "mutated", first.Formula)
	core.AssertEqual(t, "core", second.Formula)
}

func TestAX7_LoadCIConfig_Good(t *core.T) {
	dir := t.TempDir()
	core.RequireTrue(t, core.MkdirAll(core.Path(dir, ".core"), 0o755).OK)
	core.RequireTrue(t, core.WriteFile(core.Path(dir, ".core", "ci.yaml"), []byte("formula: devops\ndefault_version: v1\n"), 0o644).OK)
	t.Chdir(dir)

	cfg := LoadCIConfig()
	core.AssertEqual(t, "devops", cfg.Formula)
	core.AssertEqual(t, "v1", cfg.DefaultVersion)
}

func TestAX7_LoadCIConfig_Bad(t *core.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	cfg := LoadCIConfig()

	core.AssertEqual(t, "core", cfg.Formula)
	core.AssertEqual(t, "dev", cfg.DefaultVersion)
}

func TestAX7_LoadCIConfig_Ugly(t *core.T) {
	dir := t.TempDir()
	child := core.Path(dir, "a", "b")
	core.RequireTrue(t, core.MkdirAll(core.Path(dir, ".core"), 0o755).OK)
	core.RequireTrue(t, core.MkdirAll(child, 0o755).OK)
	core.RequireTrue(t, core.WriteFile(core.Path(dir, ".core", "ci.yaml"), []byte("tap: custom/tap\n"), 0o644).OK)
	t.Chdir(child)

	cfg := LoadCIConfig()
	core.AssertEqual(t, "custom/tap", cfg.Tap)
	core.AssertEqual(t, "core", cfg.Formula)
}

func TestAX7_LoadGitHubConfig_Good(t *core.T) {
	path := core.Path(t.TempDir(), "github.yaml")
	t.Setenv("HOOK_URL", "https://hooks.example")
	core.RequireTrue(t, core.WriteFile(path, []byte("version: 1\nlabels:\n- name: bug\n  color: ff0000\nwebhooks:\n  ci:\n    url: ${HOOK_URL}\n    events: [push]\n"), 0o644).OK)

	cfg, err := LoadGitHubConfig(path)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "https://hooks.example", cfg.Webhooks["ci"].URL)
}

func TestAX7_LoadGitHubConfig_Bad(t *core.T) {
	cfg, err := LoadGitHubConfig(core.Path(t.TempDir(), "missing.yaml"))
	core.AssertError(t, err)

	core.AssertNil(t, cfg)
	core.AssertContains(t, err.Error(), "failed to read")
}

func TestAX7_LoadGitHubConfig_Ugly(t *core.T) {
	path := core.Path(t.TempDir(), "github.yaml")
	core.RequireTrue(t, core.WriteFile(path, []byte("version: 1\nwebhooks:\n  ci:\n    url: ${MISSING_URL:-https://fallback.example}\n    events: [push]\n"), 0o644).OK)

	cfg, err := LoadGitHubConfig(path)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "https://fallback.example", cfg.Webhooks["ci"].URL)
}

func TestAX7_FindGitHubConfig_Good(t *core.T) {
	dir := t.TempDir()
	path := core.Path(dir, ".core", "github.yaml")
	core.RequireTrue(t, core.MkdirAll(core.Path(dir, ".core"), 0o755).OK)
	core.RequireTrue(t, core.WriteFile(path, []byte("version: 1\n"), 0o644).OK)

	got, err := FindGitHubConfig(dir, "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, path, got)
}

func TestAX7_FindGitHubConfig_Bad(t *core.T) {
	got, err := FindGitHubConfig(t.TempDir(), "missing.yaml")
	core.AssertError(t, err)

	core.AssertEqual(t, "", got)
	core.AssertContains(t, err.Error(), "config file not found")
}

func TestAX7_FindGitHubConfig_Ugly(t *core.T) {
	dir := t.TempDir()
	path := core.Path(dir, "github.yaml")
	core.RequireTrue(t, core.WriteFile(path, []byte("version: 1\n"), 0o644).OK)

	got, err := FindGitHubConfig(dir, "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, path, got)
}

func TestAX7_GitHubConfig_Validate_Good(t *core.T) {
	cfg := &GitHubConfig{Version: 1, Labels: []LabelConfig{{Name: "bug", Color: "ff0000"}}}
	err := cfg.Validate()

	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, cfg.Version)
}

func TestAX7_GitHubConfig_Validate_Bad(t *core.T) {
	cfg := &GitHubConfig{Version: 2}
	err := cfg.Validate()

	core.AssertError(t, err)
	core.AssertContains(t, err.Error(), "unsupported")
}

func TestAX7_GitHubConfig_Validate_Ugly(t *core.T) {
	cfg := &GitHubConfig{Version: 1, Labels: []LabelConfig{{Name: "bug", Color: "nope"}}}
	err := cfg.Validate()

	core.AssertError(t, err)
	core.AssertContains(t, err.Error(), "invalid color")
}
