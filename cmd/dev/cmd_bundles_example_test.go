package dev

import core "dappco.re/go"

func ExampleNewWorkBundle() {
	bundle, err := NewWorkBundle(WorkBundleOptions{RegistryPath: "registry.yaml"})
	core.Println(err == nil, bundle.Core != nil)
	// Output: true true
}

func ExampleWorkBundle_Start() {
	bundle, _ := NewWorkBundle(WorkBundleOptions{})
	err := bundle.Start(core.Background())
	core.Println(err == nil)
	// Output: true
}

func ExampleWorkBundle_Stop() {
	bundle, _ := NewWorkBundle(WorkBundleOptions{})
	bundle.Start(core.Background())
	err := bundle.Stop(core.Background())
	core.Println(err == nil)
	// Output: true
}
