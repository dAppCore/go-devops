package dev

import c "dappco.re/go"

func prepareExampleDevEnv(installed bool) func() {
	old := c.Getenv("CORE_IMAGES_DIR")
	dir := c.MustCast[string](c.MkdirTemp("", "dev-vm-example-*"))
	c.Setenv("CORE_IMAGES_DIR", dir)
	if installed {
		c.WriteFile(c.PathJoin(dir, vmImageName()), []byte("image"), 0o600)
	}
	return func() {
		if old == "" {
			c.Unsetenv("CORE_IMAGES_DIR")
		} else {
			c.Setenv("CORE_IMAGES_DIR", old)
		}
		c.RemoveAll(dir)
	}
}

func ExampleDevEnv_IsInstalled() {
	cleanup := prepareExampleDevEnv(true)
	defer cleanup()
	env := &DevEnv{}
	c.Println(env.IsInstalled())
	// Output: true
}

func ExampleDevEnv_Install() {
	cleanup := prepareExampleDevEnv(false)
	defer cleanup()
	env := &DevEnv{}
	r := env.Install(c.Background(), func(downloaded, total int64) {})
	c.Println(r.OK)
	// Output: false
}

func ExampleDevEnv_Boot() {
	cleanup := prepareExampleDevEnv(true)
	defer cleanup()
	env := &DevEnv{}
	r := env.Boot(c.Background(), defaultVMBootOptions())
	c.Println(r.OK)
	// Output: false
}

func ExampleDevEnv_Stop() {
	cleanup := prepareExampleDevEnv(false)
	defer cleanup()
	env := &DevEnv{}
	r := env.Stop(c.Background())
	c.Println(r.OK)
	// Output: false
}

func ExampleDevEnv_IsRunning() {
	cleanup := prepareExampleDevEnv(false)
	defer cleanup()
	env := &DevEnv{}
	running, r := env.IsRunning(c.Background())
	c.Println(r.OK, running)
	// Output: true false
}

func ExampleDevEnv_Status() {
	cleanup := prepareExampleDevEnv(true)
	defer cleanup()
	env := &DevEnv{}
	status, r := env.Status(c.Background())
	c.Println(r.OK, status.Installed, status.SSHPort)
	// Output: true true 2222
}

func ExampleDevEnv_Shell() {
	cleanup := prepareExampleDevEnv(false)
	defer cleanup()
	env := &DevEnv{}
	r := env.Shell(c.Background(), vmShellOptions{Command: []string{"echo", "ok"}})
	c.Println(r.OK)
	// Output: false
}

func ExampleDevEnv_Serve() {
	cleanup := prepareExampleDevEnv(false)
	defer cleanup()
	env := &DevEnv{}
	r := env.Serve(c.Background(), ".", vmServeOptions{Port: 3000})
	c.Println(r.OK)
	// Output: false
}

func ExampleDevEnv_Test() {
	cleanup := prepareExampleDevEnv(false)
	defer cleanup()
	env := &DevEnv{}
	r := env.Test(c.Background(), ".", vmTestOptions{Name: "unit"})
	c.Println(r.OK)
	// Output: false
}

func ExampleDevEnv_Claude() {
	cleanup := prepareExampleDevEnv(false)
	defer cleanup()
	env := &DevEnv{}
	r := env.Claude(c.Background(), ".", vmClaudeOptions{Model: "sonnet"})
	c.Println(r.OK)
	// Output: false
}

func ExampleDevEnv_CheckUpdate() {
	cleanup := prepareExampleDevEnv(false)
	defer cleanup()
	env := &DevEnv{}
	current, latest, hasUpdate, r := env.CheckUpdate(c.Background())
	c.Println(r.OK, current, latest, hasUpdate)
	// Output: true unknown unknown false
}
