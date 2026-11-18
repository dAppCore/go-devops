package docs

import . "github.com/dappcore/go"

func ExampleRepoDocInfo() {
	info := RepoDocInfo{Name: "go-devops", HasDocs: true}
	Println(info.Name, info.HasDocs)
	// Output: go-devops true
}
