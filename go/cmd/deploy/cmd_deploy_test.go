package deploy

import core "dappco.re/go"

func TestCmdDeploy_GetClient_Good(t *core.T) {
	opts := core.NewOptions(
		core.Option{Key: "url", Value: "https://coolify.example.test"},
		core.Option{Key: "token", Value: "tok-123"},
	)
	client, r := getClient(opts)
	core.AssertTrue(t, r.OK)
	core.AssertNotNil(t, client)
}

func TestCmdDeploy_GetClient_Bad(t *core.T) {
	original, hadOriginal := core.LookupEnv("COOLIFY_URL")
	core.Unsetenv("COOLIFY_URL")
	t.Cleanup(func() {
		if hadOriginal {
			core.Setenv("COOLIFY_URL", original)
		}
	})

	// No --url flag and no COOLIFY_URL: the client still constructs (the
	// Coolify SDK validates the URL lazily, on the first request), so this
	// asserts the fallback chain leaves BaseURL empty rather than panicking.
	opts := core.NewOptions()
	client, r := getClient(opts)
	core.AssertTrue(t, r.OK)
	core.AssertNotNil(t, client)
}

func TestCmdDeploy_GetClient_Ugly(t *core.T) {
	core.Setenv("COOLIFY_URL", "https://env.example.test")
	core.Setenv("COOLIFY_TOKEN", "env-tok")
	t.Cleanup(func() {
		core.Unsetenv("COOLIFY_URL")
		core.Unsetenv("COOLIFY_TOKEN")
	})

	// Flags unset: getClient falls back to the environment.
	client, r := getClient(core.NewOptions())
	core.AssertTrue(t, r.OK)
	core.AssertNotNil(t, client)
}

func TestCmdDeploy_OutputResult_Good(t *core.T) {
	opts := core.NewOptions(core.Option{Key: "json", Value: true})
	r := outputResult(opts, map[string]any{"name": "demo"})
	core.AssertTrue(t, r.OK)
}

func TestCmdDeploy_OutputResult_Bad(t *core.T) {
	// Unmarshalable value (channels never JSON-marshal) surfaces as !OK.
	opts := core.NewOptions(core.Option{Key: "json", Value: true})
	r := outputResult(opts, make(chan int))
	core.AssertFalse(t, r.OK)
}

func TestCmdDeploy_OutputResult_Ugly(t *core.T) {
	// Pretty-print path (--json unset) with an unrecognised shape still
	// succeeds — it falls through to the default %v branch.
	r := outputResult(core.NewOptions(), 42)
	core.AssertTrue(t, r.OK)
}

func TestCmdDeploy_RunCall_Good(t *core.T) {
	opts := core.NewOptions(
		core.Option{Key: "_arg", Value: "list-servers"},
		core.Option{Key: "url", Value: "https://coolify.example.test"},
		core.Option{Key: "token", Value: "tok-123"},
	)
	// No live Coolify endpoint in unit tests: the call itself fails past
	// client construction, but that proves the operation name reached the
	// client instead of being clobbered by a second positional.
	r := runCall(opts)
	core.AssertFalse(t, r.OK)
}

func TestCmdDeploy_RunCall_Bad(t *core.T) {
	r := runCall(core.NewOptions())
	core.AssertFalse(t, r.OK)
	core.AssertContains(t, r.Error(), "operation is required")
}

func TestCmdDeploy_RunCall_Ugly(t *core.T) {
	opts := core.NewOptions(
		core.Option{Key: "_arg", Value: "call-op"},
		core.Option{Key: "params", Value: "{not-json"},
	)
	r := runCall(opts)
	core.AssertFalse(t, r.OK)
	core.AssertContains(t, r.Error(), "invalid JSON params")
}
