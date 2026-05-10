// SPDX-License-Identifier: EUPL-1.2

// Service registration for the coolify package — exposes the canonical
// `NewService(opts)` + `Register(c)` shape per Mantis #1336, wrapping
// the existing `NewClient(Config) (*Client, core.Result)` constructor
// in a Core-registerable factory.
//
//	c, _ := core.New(
//	    core.WithService(coolify.NewService(coolify.ServiceOptions{
//	        Config: coolify.DefaultConfig(),
//	    })),
//	)
//	svc := core.MustServiceFor[*coolify.Service](c, "coolify")
//	servers, _ := svc.Client.ListServers(ctx)
//
// The *Client type does the heavy lifting — Service is a thin
// Core-bound handle that gives the package a registerable identity.
// Empty BaseURL leaves Client nil so consumers that boot credentials
// lazily (e.g. read from config-after-registration) can assign
// svc.Client themselves.

package coolify

import (
	core "dappco.re/go"
)

// ServiceOptions configures the coolify service. Wraps Config so callers
// can pass DefaultConfig() (env-driven) or a fully populated Config
// without a second import.
//
//	coolify.ServiceOptions{
//	    Config: coolify.DefaultConfig(),
//	}
type ServiceOptions struct {
	// Config is the Coolify connection config. Empty BaseURL → svc.Client
	// is nil; callers must wire credentials before use.
	Config Config
}

// Service is the registerable handle for the coolify package — embeds
// *core.ServiceRuntime[ServiceOptions] for typed options access and
// holds the constructed *Client.
//
// Usage example: `svc := core.MustServiceFor[*coolify.Service](c, "coolify"); _, _ = svc.Client.ListServers(ctx)`
type Service struct {
	*core.ServiceRuntime[ServiceOptions]
	// Client is the live Coolify API client. nil if NewService was
	// called with an empty Config.BaseURL — callers can construct one
	// later via NewClient(cfg) and assign to svc.Client.
	Client *Client
}

// NewService returns a factory that constructs a *Client from the
// supplied Config and wraps it as a Core-registerable *Service. If
// Config.BaseURL is empty the service is registered with a nil Client
// (no error) — callers that wire credentials lazily can mutate
// svc.Client themselves.
//
//	core.WithService(coolify.NewService(coolify.ServiceOptions{
//	    Config: coolify.DefaultConfig(),
//	}))
func NewService(opts ServiceOptions) func(*core.Core) core.Result {
	return func(c *core.Core) core.Result {
		svc := &Service{
			ServiceRuntime: core.NewServiceRuntime(c, opts),
		}
		if opts.Config.BaseURL == "" {
			return core.Ok(svc)
		}
		client, r := NewClient(opts.Config)
		if !r.OK {
			cause, _ := r.Value.(error)
			return core.Fail(core.E("coolify.NewService", "client construction failed", cause))
		}
		svc.Client = client
		return core.Ok(svc)
	}
}

// Register wires the coolify service into the Core with empty
// ServiceOptions — the imperative-style alternative to NewService.
// The resulting *Service holds a nil Client, so consumers must wire
// one via NewClient(cfg) and assign to svc.Client before use.
//
//	core.New(core.WithService(coolify.Register))
func Register(c *core.Core) core.Result {
	return NewService(ServiceOptions{})(c)
}
