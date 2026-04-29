package dev

import core "dappco.re/go"

func ExampleForgeIssue() {
	issue := ForgeIssue{Number: 42, Title: "Document release"}
	core.Println(issue.Number, issue.Title)
	// Output: 42 Document release
}
