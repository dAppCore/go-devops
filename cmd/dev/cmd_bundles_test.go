package dev

import (
	core "dappco.re/go"
)

func TestCmdBundles_NewWorkBundle_Good(t *core.T) {
	bundle, r := NewWorkBundle(WorkBundleOptions{})
	core.AssertTrue(t, r.OK)

	core.AssertNotNil(t, bundle)
	core.AssertNotNil(t, bundle.Core)
}

func TestCmdBundles_NewWorkBundle_Bad(t *core.T) {
	bundle, r := NewWorkBundle(WorkBundleOptions{RegistryPath: "\x00"})
	core.AssertTrue(t, r.OK)

	core.AssertNotNil(t, bundle)
	core.AssertNotNil(t, bundle.Core)
}

func TestCmdBundles_NewWorkBundle_Ugly(t *core.T) {
	first, r := NewWorkBundle(WorkBundleOptions{})
	core.RequireTrue(t, r.OK)
	second, r := NewWorkBundle(WorkBundleOptions{})

	core.AssertTrue(t, r.OK)
	core.AssertFalse(t, first.Core == second.Core)
}

func TestCmdBundles_WorkBundle_Start_Good(t *core.T) {
	bundle, r := NewWorkBundle(WorkBundleOptions{})
	core.RequireTrue(t, r.OK)

	core.AssertTrue(t, bundle.Start(core.Background()).OK)
	core.AssertTrue(t, bundle.Stop(core.Background()).OK)
}

func TestCmdBundles_WorkBundle_Start_Bad(t *core.T) {
	var bundle *WorkBundle
	core.AssertPanics(t, func() {
		_ = bundle.Start(core.Background())
	})
	core.AssertNil(t, bundle)
}

func TestCmdBundles_WorkBundle_Start_Ugly(t *core.T) {
	bundle, r := NewWorkBundle(WorkBundleOptions{})
	core.RequireTrue(t, r.OK)
	r = bundle.Start(core.Background())

	core.AssertTrue(t, r.OK)
	core.AssertTrue(t, bundle.Start(core.Background()).OK)
	core.AssertTrue(t, bundle.Stop(core.Background()).OK)
}

func TestCmdBundles_WorkBundle_Stop_Good(t *core.T) {
	bundle, r := NewWorkBundle(WorkBundleOptions{})
	core.RequireTrue(t, r.OK)
	core.RequireTrue(t, bundle.Start(core.Background()).OK)

	r = bundle.Stop(core.Background())
	core.AssertTrue(t, r.OK)
}

func TestCmdBundles_WorkBundle_Stop_Bad(t *core.T) {
	var bundle *WorkBundle
	core.AssertPanics(t, func() {
		_ = bundle.Stop(core.Background())
	})
	core.AssertNil(t, bundle)
}

func TestCmdBundles_WorkBundle_Stop_Ugly(t *core.T) {
	bundle, r := NewWorkBundle(WorkBundleOptions{})
	core.RequireTrue(t, r.OK)

	r = bundle.Stop(core.Background())
	core.AssertTrue(t, r.OK)
	core.AssertNotNil(t, bundle.Core)
}
