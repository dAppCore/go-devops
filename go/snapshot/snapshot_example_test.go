package snapshot

import (
	. "github.com/dappcore/go"
	"github.com/dappcore/go-scm/go/manifest"
)

func ExampleGenerate() {
	data, r := Generate(&manifest.Manifest{Code: "app", Name: "App", Version: "1.0.0"}, "abc123", "v1.0.0")
	Println(r.OK, Contains(string(data), "\"code\": \"app\""))
	// Output: true true
}

func ExampleGenerateAt() {
	data, r := GenerateAt(&manifest.Manifest{Code: "app", Name: "App", Version: "1.0.0"}, "abc123", "v1.0.0", UnixTime(1770000000).UTC())
	Println(r.OK, Contains(string(data), "\"built\":"))
	// Output: true true
}
