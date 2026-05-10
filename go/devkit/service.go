// SPDX-License-Identifier: EUPL-1.2

// Service registration for the devkit package — exposes the canonical
// `NewService(opts)` + `Register(c)` shape per Mantis #1336, wrapping
// the existing CoverageStore in a Core-registerable factory.
//
//	c, _ := core.New(
//	    core.WithService(devkit.NewService(devkit.ServiceOptions{
//	        CoverageStorePath: "/var/lib/core/coverage.json",
//	    })),
//	)
//	svc := core.MustServiceFor[*devkit.Service](c, "devkit")
//	r := svc.Coverage.Append(snapshot)
//
// The CoverageStore is the only stateful type in this package — secret
// scanning (ScanDir) is stateless and exposed directly. Service is a
// thin Core-bound handle that gives the package a registerable identity
// the rest of the framework can discover via core.ServiceFor.

package devkit

import (
	core "dappco.re/go"
)

// ServiceOptions configures the devkit service. Empty CoverageStorePath
// leaves Coverage nil — callers that wire the store path lazily can
// assign svc.Coverage themselves.
//
//	devkit.ServiceOptions{
//	    CoverageStorePath: "/var/lib/core/coverage.json",
//	}
type ServiceOptions struct {
	// CoverageStorePath is the on-disk path for the persistent coverage
	// snapshot store. Empty → svc.Coverage is nil; callers can wire
	// later via NewCoverageStore(path).
	CoverageStorePath string
}

// Service is the registerable handle for the devkit package — embeds
// *core.ServiceRuntime[ServiceOptions] for typed options access and
// holds the constructed *CoverageStore.
//
// Usage example: `svc := core.MustServiceFor[*devkit.Service](c, "devkit"); _ = svc.Coverage.Append(snapshot)`
type Service struct {
	*core.ServiceRuntime[ServiceOptions]
	// Coverage is the persistent coverage snapshot store. nil if
	// NewService was called with an empty CoverageStorePath.
	Coverage *CoverageStore
}

// NewService returns a factory that constructs a *CoverageStore from
// the supplied path and wraps it as a Core-registerable *Service.
// Empty CoverageStorePath registers the service with a nil Coverage —
// callers can mutate svc.Coverage later.
//
//	core.WithService(devkit.NewService(devkit.ServiceOptions{
//	    CoverageStorePath: "/var/lib/core/coverage.json",
//	}))
func NewService(opts ServiceOptions) func(*core.Core) core.Result {
	return func(c *core.Core) core.Result {
		svc := &Service{
			ServiceRuntime: core.NewServiceRuntime(c, opts),
		}
		if opts.CoverageStorePath != "" {
			svc.Coverage = NewCoverageStore(opts.CoverageStorePath)
		}
		return core.Ok(svc)
	}
}

// Register wires the devkit service into the Core with empty
// ServiceOptions — the imperative-style alternative to NewService.
// The resulting *Service holds a nil Coverage, so consumers must
// either assign svc.Coverage themselves or boot via NewService with
// a CoverageStorePath.
//
//	core.New(core.WithService(devkit.Register))
func Register(c *core.Core) core.Result {
	return NewService(ServiceOptions{})(c)
}
