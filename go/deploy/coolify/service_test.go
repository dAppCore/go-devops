package coolify

import (
	core "dappco.re/go"
)

// TestService_NewService_Good registers the coolify service with an empty
// Config (no BaseURL) and verifies the factory yields a *Service with a
// nil Client and no error — the lazy-credential path.
func TestService_NewService_Good(t *core.T) {
	factory := NewService(ServiceOptions{})
	c := core.New()
	r := factory(c)

	core.RequireTrue(t, r.OK)
	svc := core.MustCast[*Service](r)
	core.AssertNotNil(t, svc)
	core.AssertNil(t, svc.Client)
}

// TestService_NewService_Bad confirms a populated BaseURL drives the
// NewClient construction path; with Python init stubbed to fail the
// factory surfaces the wrapped construction error.
func TestService_NewService_Bad(t *core.T) {
	stubCoolifyInit(t, core.AnError)
	factory := NewService(ServiceOptions{Config: Config{BaseURL: "https://coolify.example", APIToken: "secret"}})
	c := core.New()
	r := factory(c)

	core.AssertFalse(t, r.OK)
	core.AssertContains(t, r.Error(), "client construction failed")
}

// TestService_NewService_Ugly drives the full happy construction path:
// a valid Config with Python init stubbed to succeed yields a *Service
// whose Client is wired.
func TestService_NewService_Ugly(t *core.T) {
	stubCoolifyInit(t, nil)
	factory := NewService(ServiceOptions{Config: Config{BaseURL: "https://coolify.example", APIToken: "secret", Timeout: 5}})
	c := core.New()
	r := factory(c)

	core.RequireTrue(t, r.OK)
	svc := core.MustCast[*Service](r)
	core.AssertNotNil(t, svc.Client)
	core.AssertEqual(t, "https://coolify.example", svc.Client.baseURL)
}

// TestService_Register_Good verifies the imperative Register entrypoint
// registers a *Service with a nil Client (empty ServiceOptions).
func TestService_Register_Good(t *core.T) {
	c := core.New()
	r := Register(c)

	core.RequireTrue(t, r.OK)
	svc := core.MustCast[*Service](r)
	core.AssertNotNil(t, svc)
	core.AssertNil(t, svc.Client)
}

// TestService_Register_Discoverable confirms the service is reachable via
// the Core service registry after registration through WithService.
func TestService_Register_Discoverable(t *core.T) {
	c := core.New(core.WithService(Register))

	names := c.Services()
	core.AssertGreater(t, len(names), 1)
}
