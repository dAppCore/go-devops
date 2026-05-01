package dev

import core "dappco.re/go"

func TestService_ServiceOptions_Good(t *core.T) {
	opts := ServiceOptions{RegistryPath: "repos.yaml"}

	core.AssertEqual(t, "repos.yaml", opts.RegistryPath)
	core.AssertNotEmpty(t, opts.RegistryPath)
}

func TestService_ServiceOptions_Bad(t *core.T) {
	opts := ServiceOptions{}

	core.AssertEqual(t, "", opts.RegistryPath)
	core.AssertEmpty(t, opts.RegistryPath)
}

func TestService_ServiceOptions_Ugly(t *core.T) {
	opts := ServiceOptions{RegistryPath: " spaced path.yaml "}

	core.AssertContains(t, opts.RegistryPath, "spaced")
	core.AssertNotEqual(t, "spaced path.yaml", opts.RegistryPath)
}

func TestService_Service_Good(t *core.T) {
	svc := &Service{}

	core.AssertNotNil(t, svc)
	core.AssertNil(t, svc.ServiceRuntime)
}

func TestService_Service_Bad(t *core.T) {
	var svc *Service

	core.AssertNil(t, svc)
	core.AssertPanics(t, func() { _ = svc.ServiceRuntime })
}

func TestService_Service_Ugly(t *core.T) {
	svc := Service{}

	core.AssertNil(t, svc.ServiceRuntime)
	core.AssertEqual(t, Service{}, svc)
}
