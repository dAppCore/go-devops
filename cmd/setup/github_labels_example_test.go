package setup

import . "dappco.re/go"

func labelsExampleFakeGH(body string) func() {
	dir := MustCast[string](MkdirTemp("", "labels-gh-*"))
	WriteFile(PathJoin(dir, "gh"), []byte("#!/bin/sh\n"+body+"\n"), 0o755)
	oldPath := Getenv("PATH")
	Setenv("PATH", dir+":"+oldPath)
	return func() {
		Setenv("PATH", oldPath)
		RemoveAll(dir)
	}
}

func ExampleListLabels() {
	cleanup := labelsExampleFakeGH("echo '[{\"name\":\"bug\",\"color\":\"ff0000\",\"description\":\"Bug\"}]'")
	defer cleanup()
	labels, err := ListLabels("owner/repo")
	Println(err == nil, labels[0].Name)
	// Output: true bug
}

func ExampleCreateLabel() {
	cleanup := labelsExampleFakeGH("exit 0")
	defer cleanup()
	err := CreateLabel("owner/repo", LabelConfig{Name: "bug", Color: "ff0000"})
	Println(err == nil)
	// Output: true
}

func ExampleEditLabel() {
	cleanup := labelsExampleFakeGH("exit 0")
	defer cleanup()
	err := EditLabel("owner/repo", LabelConfig{Name: "bug", Color: "00ff00"})
	Println(err == nil)
	// Output: true
}

func ExampleSyncLabels() {
	cleanup := labelsExampleFakeGH("echo '[{\"name\":\"bug\",\"color\":\"00ff00\",\"description\":\"old\"}]'")
	defer cleanup()
	changes, err := SyncLabels("owner/repo", &GitHubConfig{Labels: []LabelConfig{{Name: "bug", Color: "ff0000"}}}, true)
	Println(err == nil, changes.HasChanges())
	// Output: true true
}
