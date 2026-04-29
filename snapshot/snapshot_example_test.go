package snapshot

import (
	. "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

func ExampleGenerate() {
	data, err := Generate(&manifest.Manifest{Code: "app", Name: "App", Version: "1.0.0"}, "abc123", "v1.0.0")
	Println(err == nil, Contains(string(data), "\"code\": \"app\""))
	// Output: true true
}

func ExampleGenerateAt() {
	data, err := GenerateAt(&manifest.Manifest{Code: "app", Name: "App", Version: "1.0.0"}, "abc123", "v1.0.0", UnixTime(1770000000).UTC())
	Println(err == nil, Contains(string(data), "\"built\":"))
	// Output: true true
}
