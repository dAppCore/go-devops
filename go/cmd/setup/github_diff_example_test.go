package setup

import . "dappco.re/go"

func ExampleNewChangeSet() {
	cs := NewChangeSet("owner/repo")
	Println(cs.Repo, len(cs.Changes))
	// Output: owner/repo 0
}

func ExampleChangeSet_Add() {
	cs := NewChangeSet("owner/repo")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "create label")
	Println(cs.Changes[0].Name, cs.Changes[0].Type)
	// Output: bug create
}

func ExampleChangeSet_AddWithDetails() {
	cs := NewChangeSet("owner/repo")
	cs.AddWithDetails(CategoryLabel, ChangeUpdate, "bug", "update", map[string]string{"color": "old -> new"})
	Println(cs.Changes[0].Details["color"])
	// Output: old -> new
}

func ExampleChangeSet_HasChanges() {
	cs := NewChangeSet("owner/repo")
	cs.Add(CategoryLabel, ChangeSkip, "bug", "up to date")
	Println(cs.HasChanges())
	// Output: false
}

func ExampleChangeSet_Count() {
	cs := NewChangeSet("owner/repo")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "")
	creates, updates, deletes, skips := cs.Count()
	Println(creates, updates, deletes, skips)
	// Output: 1 0 0 0
}

func ExampleChangeSet_CountByCategory() {
	cs := NewChangeSet("owner/repo")
	cs.Add(CategoryWebhook, ChangeCreate, "ci", "")
	Println(cs.CountByCategory()[CategoryWebhook])
	// Output: 1
}

func ExampleChangeSet_Print() {
	cs := NewChangeSet("owner/repo")
	Println(cs.HasChanges())
	// Output: false
}

func ExampleNewAggregate() {
	agg := NewAggregate()
	Println(len(agg.Sets))
	// Output: 0
}

func ExampleAggregate_Add() {
	agg := NewAggregate()
	agg.Add(NewChangeSet("owner/repo"))
	Println(len(agg.Sets))
	// Output: 1
}

func ExampleAggregate_TotalChanges() {
	cs := NewChangeSet("owner/repo")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "")
	agg := NewAggregate()
	agg.Add(cs)
	creates, updates, deletes, skips := agg.TotalChanges()
	Println(creates, updates, deletes, skips)
	// Output: 1 0 0 0
}

func ExampleAggregate_ReposWithChanges() {
	cs := NewChangeSet("owner/repo")
	cs.Add(CategoryLabel, ChangeCreate, "bug", "")
	agg := NewAggregate()
	agg.Add(cs)
	Println(agg.ReposWithChanges())
	// Output: 1
}

func ExampleAggregate_PrintSummary() {
	agg := NewAggregate()
	Println(agg.ReposWithChanges())
	// Output: 0
}
