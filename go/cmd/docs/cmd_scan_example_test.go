package docs

import . "dappco.re/go"

func ExampleRepoDocInfo() {
	info := RepoDocInfo{Name: "go-devops", HasDocs: true}
	Println(info.Name, info.HasDocs)
	// Output: go-devops true
}
