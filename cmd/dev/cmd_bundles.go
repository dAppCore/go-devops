package dev

import (
	"context"
	"fmt"

	"dappco.re/go/core"
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
func NewWorkBundle(opts WorkBundleOptions) (*WorkBundle, error) {
	c := core.New()

	svc := &Service{
		ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{
			RegistryPath: opts.RegistryPath,
		}),
	}

	c.Service("dev", core.Service{
		OnStart: func() core.Result {
			c.RegisterTask(svc.handleTask)
			return core.Result{OK: true}
		},
	})

	c.LockEnable()
	c.LockApply()

	return &WorkBundle{Core: c}, nil
}

// Start initialises the bundle services.
func (b *WorkBundle) Start(ctx context.Context) error {
	return resultError(b.Core.ServiceStartup(ctx, nil))
}

// Stop shuts down the bundle services.
func (b *WorkBundle) Stop(ctx context.Context) error {
	return resultError(b.Core.ServiceShutdown(ctx))
}

// resultError extracts an error from a failed core.Result, returning nil on success.
func resultError(r core.Result) error {
	if !r.OK {
		if err, ok := r.Value.(error); ok {
			return err
		}
		if r.Value != nil {
			return fmt.Errorf("service operation failed: %v", r.Value)
		}
		return fmt.Errorf("service operation failed")
	}
	return nil
}
