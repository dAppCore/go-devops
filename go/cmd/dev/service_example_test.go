package dev

import core "dappco.re/go"

func ExampleServiceOptions() {
	opts := ServiceOptions{RegistryPath: "repos.yaml"}
	core.Println(opts.RegistryPath)
	// Output: repos.yaml
}

func ExampleService() {
	svc := &Service{}
	core.Println(svc.ServiceRuntime == nil)
	// Output: true
}
