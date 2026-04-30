package dev

import (
	"context"

	core "dappco.re/go"
)

// WorkBundle contains the Core instance for dev work operations.
type WorkBundle struct {
	Core *core.Core
}

// WorkBundleOptions configures the work bundle.
type WorkBundleOptions struct {
	RegistryPath string
}

// NewWorkBundle creates a bundle for dev work operations.
// Includes: dev (orchestration) service.
func NewWorkBundle(opts WorkBundleOptions) (*WorkBundle, core.Result) {
	c := core.New()

	svc := &Service{
		ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{
			RegistryPath: opts.RegistryPath,
		}),
	}

	c.Service("dev", core.Service{
		OnStart: func() core.Result {
			c.RegisterAction(svc.handleAction)
			return core.Ok(nil)
		},
	})

	c.LockEnable()
	c.LockApply()

	return &WorkBundle{Core: c}, core.Ok(nil)
}

// Start initialises the bundle services.
func (b *WorkBundle) Start(ctx context.Context) (_ core.Result) {
	return b.Core.ServiceStartup(ctx, nil)
}

// Stop shuts down the bundle services.
func (b *WorkBundle) Stop(ctx context.Context) (_ core.Result) {
	return b.Core.ServiceShutdown(ctx)
}

// resultError extracts an error from a failed core.Result, returning nil on success.
func resultError(r core.Result) (_ core.Result) {
	if !r.OK {
		if err, ok := r.Value.(error); ok {
			return core.Fail(err)
		}
		if r.Value != nil {
			return core.Fail(core.Errorf("service operation failed: %v", r.Value))
		}
		return core.Fail(core.Errorf("service operation failed"))
	}
	return core.Ok(nil)
}
