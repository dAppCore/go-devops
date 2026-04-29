package dev

import (
	core "dappco.re/go"
)

func TestCmdBundles_NewWorkBundle_Good(t *core.T) {
	bundle, err := NewWorkBundle(WorkBundleOptions{})
	core.AssertNoError(t, err)

	core.AssertNotNil(t, bundle)
	core.AssertNotNil(t, bundle.Core)
}

func TestCmdBundles_NewWorkBundle_Bad(t *core.T) {
	bundle, err := NewWorkBundle(WorkBundleOptions{RegistryPath: "\x00"})
	core.AssertNoError(t, err)

	core.AssertNotNil(t, bundle)
	core.AssertNotNil(t, bundle.Core)
}

func TestCmdBundles_NewWorkBundle_Ugly(t *core.T) {
	first, err := NewWorkBundle(WorkBundleOptions{})
	core.RequireNoError(t, err)
	second, err := NewWorkBundle(WorkBundleOptions{})

	core.AssertNoError(t, err)
	core.AssertFalse(t, first.Core == second.Core)
}

func TestCmdBundles_WorkBundle_Start_Good(t *core.T) {
	bundle, err := NewWorkBundle(WorkBundleOptions{})
	core.RequireNoError(t, err)

	core.AssertNoError(t, bundle.Start(core.Background()))
	core.AssertNoError(t, bundle.Stop(core.Background()))
}

func TestCmdBundles_WorkBundle_Start_Bad(t *core.T) {
	var bundle *WorkBundle
	core.AssertPanics(t, func() {
		_ = bundle.Start(core.Background())
	})
	core.AssertNil(t, bundle)
}

func TestCmdBundles_WorkBundle_Start_Ugly(t *core.T) {
	bundle, err := NewWorkBundle(WorkBundleOptions{})
	core.RequireNoError(t, err)
	err = bundle.Start(core.Background())

	core.AssertNoError(t, err)
	core.AssertNoError(t, bundle.Start(core.Background()))
	core.AssertNoError(t, bundle.Stop(core.Background()))
}

func TestCmdBundles_WorkBundle_Stop_Good(t *core.T) {
	bundle, err := NewWorkBundle(WorkBundleOptions{})
	core.RequireNoError(t, err)
	core.RequireNoError(t, bundle.Start(core.Background()))

	err = bundle.Stop(core.Background())
	core.AssertNoError(t, err)
}

func TestCmdBundles_WorkBundle_Stop_Bad(t *core.T) {
	var bundle *WorkBundle
	core.AssertPanics(t, func() {
		_ = bundle.Stop(core.Background())
	})
	core.AssertNil(t, bundle)
}

func TestCmdBundles_WorkBundle_Stop_Ugly(t *core.T) {
	bundle, err := NewWorkBundle(WorkBundleOptions{})
	core.RequireNoError(t, err)

	err = bundle.Stop(core.Background())
	core.AssertNoError(t, err)
	core.AssertNotNil(t, bundle.Core)
}
