package docs

import . "dappco.re/go"

func TestCmdScan_RepoDocInfo_Good(t *T) {
	info := RepoDocInfo{Name: "go-devops", HasDocs: true, DocsFiles: []string{"docs/index.md"}}

	AssertEqual(t, "go-devops", info.Name)
	AssertTrue(t, info.HasDocs)
}

func TestCmdScan_RepoDocInfo_Bad(t *T) {
	info := RepoDocInfo{}

	AssertEqual(t, "", info.Name)
	AssertFalse(t, info.HasDocs)
}

func TestCmdScan_RepoDocInfo_Ugly(t *T) {
	info := RepoDocInfo{KBFiles: []string{}}

	AssertEmpty(t, info.KBFiles)
	AssertEqual(t, "", info.Path)
}
