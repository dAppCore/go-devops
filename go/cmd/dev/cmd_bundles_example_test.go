package dev

import core "dappco.re/go"

func ExampleNewWorkBundle() {
	bundle, r := NewWorkBundle(WorkBundleOptions{RegistryPath: "registry.yaml"})
	core.Println(r.OK, bundle.Core != nil)
	// Output: true true
}

func ExampleWorkBundle_Start() {
	bundle, r := NewWorkBundle(WorkBundleOptions{})
	if !r.OK {
		core.Println(false)
		return
	}
	start := bundle.Start(core.Background())
	core.Println(start.OK)
	// Output: true
}

func ExampleWorkBundle_Stop() {
	bundle, r := NewWorkBundle(WorkBundleOptions{})
	if !r.OK {
		core.Println(false)
		return
	}
	start := bundle.Start(core.Background())
	if !start.OK {
		core.Println(false)
		return
	}
	stop := bundle.Stop(core.Background())
	core.Println(stop.OK)
	// Output: true
}
