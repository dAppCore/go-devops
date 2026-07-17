package dev

import (
	"context"
	"runtime"
	"time"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
	log "dappco.re/go/log"
)

const vmDefaultSSHPort = 2222

// DevEnv provides the local development VM backend used by dev VM commands.
type DevEnv struct{}

type vmBootOptions struct {
	Memory int
	CPUs   int
	Name   string
	Fresh  bool
}

type vmShellOptions struct {
	Console bool
	Command []string
}

type vmServeOptions struct {
	Port int
	Path string
}

type vmTestOptions struct {
	Name    string
	Command []string
}

type vmClaudeOptions struct {
	NoAuth bool
	Model  string
	Auth   []string
}

type vmStatus struct {
	Installed    bool
	Running      bool
	ImageVersion string
	ContainerID  string
	Memory       int
	CPUs         int
	SSHPort      int
	Uptime       time.Duration
}

func newVMDevEnv() (*DevEnv, core.Result) {
	return &DevEnv{}, core.Ok(nil)
}

func vmImageName() string {
	return core.Sprintf("core-devops-%s-%s.qcow2", runtime.GOOS, runtime.GOARCH)
}

func vmImagePath() core.Result {
	if dir := core.Getenv("CORE_IMAGES_DIR"); dir != "" {
		return core.Ok(core.PathJoin(dir, vmImageName()))
	}
	home := core.UserHomeDir()
	if !home.OK {
		return home
	}
	return core.Ok(core.PathJoin(home.Value.(string), ".core", "images", vmImageName()))
}

func defaultVMBootOptions() vmBootOptions {
	return vmBootOptions{
		Memory: 4096,
		CPUs:   2,
		Name:   "core-dev",
	}
}

func (d *DevEnv) IsInstalled() bool {
	path := vmImagePath()
	return path.OK && core.Stat(path.Value.(string)).OK
}

func (d *DevEnv) Install(context.Context, func(downloaded, total int64)) core.Result {
	return core.Fail(core.Errorf("dev VM image installer backend is unavailable"))
}

func (d *DevEnv) Boot(context.Context, vmBootOptions) core.Result {
	if !d.IsInstalled() {
		return core.Fail(log.E("dev.vm", i18n.T("cmd.dev.vm.not_installed"), nil))
	}
	return core.Fail(core.Errorf("dev VM boot backend is unavailable"))
}

func (d *DevEnv) Stop(context.Context) core.Result {
	return core.Fail(core.Errorf("dev VM stop backend is unavailable"))
}

func (d *DevEnv) IsRunning(context.Context) (bool, core.Result) {
	return false, core.Ok(nil)
}

func (d *DevEnv) Status(context.Context) (vmStatus, core.Result) {
	return vmStatus{
		Installed: d.IsInstalled(),
		SSHPort:   vmDefaultSSHPort,
		Memory:    4096,
		CPUs:      2,
	}, core.Ok(nil)
}

func (d *DevEnv) Shell(context.Context, vmShellOptions) core.Result {
	return core.Fail(core.Errorf("dev VM shell backend is unavailable"))
}

func (d *DevEnv) Serve(context.Context, string, vmServeOptions) core.Result {
	return core.Fail(core.Errorf("dev VM serve backend is unavailable"))
}

func (d *DevEnv) Test(context.Context, string, vmTestOptions) core.Result {
	return core.Fail(core.Errorf("dev VM test backend is unavailable"))
}

func (d *DevEnv) Claude(context.Context, string, vmClaudeOptions) core.Result {
	return core.Fail(core.Errorf("dev VM Claude backend is unavailable"))
}

func (d *DevEnv) CheckUpdate(context.Context) (string, string, bool, core.Result) {
	return "unknown", "unknown", false, core.Ok(nil)
}

// addVMCommands adds the dev environment VM commands under "dev".
// These are direct subcommands: core dev install, core dev boot, etc.
func addVMCommands(c *core.Core) core.Result {
	for _, register := range []func(*core.Core) core.Result{
		addVMInstallCommand,
		addVMBootCommand,
		addVMStopCommand,
		addVMStatusCommand,
		addVMShellCommand,
		addVMServeCommand,
		addVMTestCommand,
		addVMClaudeCommand,
		addVMUpdateCommand,
	} {
		if r := register(c); !r.OK {
			return r
		}
	}
	return core.Ok(nil)
}

// addVMInstallCommand adds the 'dev install' command.
func addVMInstallCommand(c *core.Core) core.Result {
	return c.Command("dev/install", core.Command{
		Description: i18n.T("cmd.dev.vm.install.short"),
		Action: func(core.Options) core.Result {
			return runVMInstall()
		},
	})
}

func runVMInstall() (_ core.Result) {
	d, r := newVMDevEnv()
	if !r.OK {
		return r
	}

	if d.IsInstalled() {
		cli.Text(successStyle.Render(i18n.T("cmd.dev.vm.already_installed")))
		cli.Blank()
		cli.Text(i18n.T("cmd.dev.vm.check_updates", map[string]any{"Command": dimStyle.Render("core dev update")}))
		return core.Ok(nil)
	}

	cli.Print("%s %s\n", dimStyle.Render(i18n.Label("image")), vmImageName())
	cli.Blank()
	cli.Text(i18n.T("cmd.dev.vm.downloading"))
	cli.Blank()

	ctx := context.Background()
	start := time.Now()
	var lastProgress int64

	r = d.Install(ctx, func(downloaded, total int64) {
		if total > 0 {
			pct := int(float64(downloaded) / float64(total) * 100)
			if pct != int(float64(lastProgress)/float64(total)*100) {
				cli.Print("\r%s %d%%", dimStyle.Render(i18n.T("cmd.dev.vm.progress_label")), pct)
				lastProgress = downloaded
			}
		}
	})

	cli.Blank() // Clear progress line

	if !r.OK {
		return cli.Wrap(r.Value.(error), "install failed")
	}

	elapsed := time.Since(start).Round(time.Second)
	cli.Blank()
	cli.Text(i18n.T("cmd.dev.vm.installed_in", map[string]any{"Duration": elapsed}))
	cli.Blank()
	cli.Text(i18n.T("cmd.dev.vm.start_with", map[string]any{"Command": dimStyle.Render("core dev boot")}))

	return core.Ok(nil)
}

// addVMBootCommand adds the 'dev boot' command.
func addVMBootCommand(c *core.Core) core.Result {
	return c.Command("dev/boot", core.Command{
		Description: i18n.T("cmd.dev.vm.boot.short"),
		Flags: core.NewOptions(
			core.Option{Key: "memory", Value: 0},
			core.Option{Key: "cpus", Value: 0},
			core.Option{Key: "fresh", Value: false},
		),
		Action: func(o core.Options) core.Result {
			return runVMBoot(o.Int("memory"), o.Int("cpus"), o.Bool("fresh"))
		},
	})
}

func runVMBoot(memory, cpus int, fresh bool) (_ core.Result) {
	d, r := newVMDevEnv()
	if !r.OK {
		return r
	}

	if !d.IsInstalled() {
		return core.Fail(log.E("dev.vm", i18n.T("cmd.dev.vm.not_installed"), nil))
	}

	opts := defaultVMBootOptions()
	if memory > 0 {
		opts.Memory = memory
	}
	if cpus > 0 {
		opts.CPUs = cpus
	}
	opts.Fresh = fresh

	cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.config_label")), i18n.T("cmd.dev.vm.config_value", map[string]any{"Memory": opts.Memory, "CPUs": opts.CPUs}))
	cli.Blank()
	cli.Text(i18n.T("cmd.dev.vm.booting"))

	ctx := context.Background()
	if r := d.Boot(ctx, opts); !r.OK {
		return r
	}

	cli.Blank()
	cli.Text(successStyle.Render(i18n.T("cmd.dev.vm.running")))
	cli.Blank()
	cli.Text(i18n.T("cmd.dev.vm.connect_with", map[string]any{"Command": dimStyle.Render("core dev shell")}))
	cli.Print("%s %s\n", i18n.T("cmd.dev.vm.ssh_port"), dimStyle.Render("2222"))

	return core.Ok(nil)
}

// addVMStopCommand adds the 'dev stop' command.
func addVMStopCommand(c *core.Core) core.Result {
	return c.Command("dev/stop", core.Command{
		Description: i18n.T("cmd.dev.vm.stop.short"),
		Action: func(core.Options) core.Result {
			return runVMStop()
		},
	})
}

func runVMStop() (_ core.Result) {
	d, r := newVMDevEnv()
	if !r.OK {
		return r
	}

	ctx := context.Background()
	running, r := d.IsRunning(ctx)
	if !r.OK {
		return r
	}

	if !running {
		cli.Text(dimStyle.Render(i18n.T("cmd.dev.vm.not_running")))
		return core.Ok(nil)
	}

	cli.Text(i18n.T("cmd.dev.vm.stopping"))

	if r := d.Stop(ctx); !r.OK {
		return r
	}

	cli.Text(successStyle.Render(i18n.T("common.status.stopped")))
	return core.Ok(nil)
}

// addVMStatusCommand adds the 'dev status' command, plus a 'dev vm-status'
// alias — core.Command has no native Aliases field, so the alias is a
// second registration of the same Action (mirrors dappco.re/go/agent's
// multi-path registration pattern for command aliases).
func addVMStatusCommand(c *core.Core) core.Result {
	action := func(core.Options) core.Result { return runVMStatus() }
	if r := c.Command("dev/status", core.Command{
		Description: i18n.T("cmd.dev.vm.status.short"),
		Action:      action,
	}); !r.OK {
		return r
	}
	return c.Command("dev/vm-status", core.Command{
		Description: i18n.T("cmd.dev.vm.status.short"),
		Action:      action,
	})
}

func runVMStatus() (_ core.Result) {
	d, r := newVMDevEnv()
	if !r.OK {
		return r
	}

	ctx := context.Background()
	status, r := d.Status(ctx)
	if !r.OK {
		return r
	}

	cli.Text(headerStyle.Render(i18n.T("cmd.dev.vm.status_title")))
	cli.Blank()

	// Installation status
	if status.Installed {
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.installed_label")), successStyle.Render(i18n.T("cmd.dev.vm.installed_yes")))
		if status.ImageVersion != "" {
			cli.Print("%s %s\n", dimStyle.Render(i18n.Label("version")), status.ImageVersion)
		}
	} else {
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.installed_label")), errorStyle.Render(i18n.T("cmd.dev.vm.installed_no")))
		cli.Blank()
		cli.Text(i18n.T("cmd.dev.vm.install_with", map[string]any{"Command": dimStyle.Render("core dev install")}))
		return core.Ok(nil)
	}

	cli.Blank()

	// Running status
	if status.Running {
		cli.Print("%s %s\n", dimStyle.Render(i18n.Label("status")), successStyle.Render(i18n.T("common.status.running")))
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.container_label")), status.ContainerID[:8])
		cli.Print("%s %dMB\n", dimStyle.Render(i18n.T("cmd.dev.vm.memory_label")), status.Memory)
		cli.Print("%s %d\n", dimStyle.Render(i18n.T("cmd.dev.vm.cpus_label")), status.CPUs)
		cli.Print("%s %d\n", dimStyle.Render(i18n.T("cmd.dev.vm.ssh_port")), status.SSHPort)
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.uptime_label")), formatVMUptime(status.Uptime))
	} else {
		cli.Print("%s %s\n", dimStyle.Render(i18n.Label("status")), dimStyle.Render(i18n.T("common.status.stopped")))
		cli.Blank()
		cli.Text(i18n.T("cmd.dev.vm.start_with", map[string]any{"Command": dimStyle.Render("core dev boot")}))
	}

	return core.Ok(nil)
}

func formatVMUptime(d time.Duration) string {
	if d < time.Minute {
		return cli.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return cli.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return cli.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	return cli.Sprintf("%dd %dh", int(d.Hours()/24), int(d.Hours())%24)
}

// addVMShellCommand adds the 'dev shell [-- command...]' command. The
// backend (DevEnv.Shell) is an unconditional stub ("backend is unavailable")
// pending a real VM implementation, so the historical `-- command...`
// passthrough — which needs multiple trailing positionals that
// core.Cli.Run cannot carry (only the last bare positional survives under
// "_arg") — is reduced to a single positional for now. Revisit once a real
// backend lands and the full command line needs to reach it.
func addVMShellCommand(c *core.Core) core.Result {
	return c.Command("dev/shell", core.Command{
		Description: i18n.T("cmd.dev.vm.shell.short"),
		Flags: core.NewOptions(
			core.Option{Key: "console", Value: false},
		),
		Action: func(o core.Options) core.Result {
			var command []string
			if arg := o.String("_arg"); arg != "" {
				command = []string{arg}
			}
			return runVMShell(o.Bool("console"), command)
		},
	})
}

func runVMShell(console bool, command []string) (_ core.Result) {
	d, r := newVMDevEnv()
	if !r.OK {
		return r
	}

	opts := vmShellOptions{
		Console: console,
		Command: command,
	}

	ctx := context.Background()
	return d.Shell(ctx, opts)
}

// addVMServeCommand adds the 'dev serve' command.
func addVMServeCommand(c *core.Core) core.Result {
	return c.Command("dev/serve", core.Command{
		Description: i18n.T("cmd.dev.vm.serve.short"),
		Flags: core.NewOptions(
			core.Option{Key: "port", Value: 0},
			core.Option{Key: "path", Value: ""},
		),
		Action: func(o core.Options) core.Result {
			return runVMServe(o.Int("port"), o.String("path"))
		},
	})
}

func runVMServe(port int, path string) (_ core.Result) {
	d, r := newVMDevEnv()
	if !r.OK {
		return r
	}

	projectDirResult := core.Getwd()
	if !projectDirResult.OK {
		return projectDirResult
	}
	projectDir := projectDirResult.Value.(string)

	opts := vmServeOptions{
		Port: port,
		Path: path,
	}

	ctx := context.Background()
	return d.Serve(ctx, projectDir, opts)
}

// addVMTestCommand adds the 'dev test [-- command...]' command. Same
// single-positional reduction as addVMShellCommand — see its comment; the
// backend (DevEnv.Test) is likewise an unconditional stub.
func addVMTestCommand(c *core.Core) core.Result {
	return c.Command("dev/test", core.Command{
		Description: i18n.T("cmd.dev.vm.test.short"),
		Flags: core.NewOptions(
			core.Option{Key: "name", Value: ""},
		),
		Action: func(o core.Options) core.Result {
			var command []string
			if arg := o.String("_arg"); arg != "" {
				command = []string{arg}
			}
			return runVMTest(o.String("name"), command)
		},
	})
}

func runVMTest(name string, command []string) (_ core.Result) {
	d, r := newVMDevEnv()
	if !r.OK {
		return r
	}

	projectDirResult := core.Getwd()
	if !projectDirResult.OK {
		return projectDirResult
	}
	projectDir := projectDirResult.Value.(string)

	opts := vmTestOptions{
		Name:    name,
		Command: command,
	}

	ctx := context.Background()
	return d.Test(ctx, projectDir, opts)
}

// addVMClaudeCommand adds the 'dev claude' command. --auth was a cobra
// StringSlice (repeatable); core.Options only keeps the last value for a
// repeated flag key, so --auth now takes a single comma-separated value
// (e.g. --auth=token,cookie). The backend (DevEnv.Claude) is an
// unconditional stub, so this is a documentation-level change for now.
func addVMClaudeCommand(c *core.Core) core.Result {
	return c.Command("dev/claude", core.Command{
		Description: i18n.T("cmd.dev.vm.claude.short"),
		Flags: core.NewOptions(
			core.Option{Key: "no-auth", Value: false},
			core.Option{Key: "model", Value: ""},
			core.Option{Key: "auth", Value: ""},
		),
		Action: func(o core.Options) core.Result {
			var authFlags []string
			for _, a := range core.Split(o.String("auth"), ",") {
				if a = core.Trim(a); a != "" {
					authFlags = append(authFlags, a)
				}
			}
			return runVMClaude(o.Bool("no-auth"), o.String("model"), authFlags)
		},
	})
}

func runVMClaude(noAuth bool, model string, authFlags []string) (_ core.Result) {
	d, r := newVMDevEnv()
	if !r.OK {
		return r
	}

	projectDirResult := core.Getwd()
	if !projectDirResult.OK {
		return projectDirResult
	}
	projectDir := projectDirResult.Value.(string)

	opts := vmClaudeOptions{
		NoAuth: noAuth,
		Model:  model,
		Auth:   authFlags,
	}

	ctx := context.Background()
	return d.Claude(ctx, projectDir, opts)
}

// addVMUpdateCommand adds the 'dev update' command.
func addVMUpdateCommand(c *core.Core) core.Result {
	return c.Command("dev/update", core.Command{
		Description: i18n.T("cmd.dev.vm.update.short"),
		Flags: core.NewOptions(
			core.Option{Key: "apply", Value: false},
		),
		Action: func(o core.Options) core.Result {
			return runVMUpdate(o.Bool("apply"))
		},
	})
}

func runVMUpdate(apply bool) (_ core.Result) {
	d, r := newVMDevEnv()
	if !r.OK {
		return r
	}

	ctx := context.Background()

	cli.Text(i18n.T("common.progress.checking_updates"))
	cli.Blank()

	current, latest, hasUpdate, r := d.CheckUpdate(ctx)
	if !r.OK {
		return cli.Wrap(r.Value.(error), "failed to check for updates")
	}

	cli.Print("%s %s\n", dimStyle.Render(i18n.Label("current")), valueStyle.Render(current))
	cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.latest_label")), valueStyle.Render(latest))
	cli.Blank()

	if !hasUpdate {
		cli.Text(successStyle.Render(i18n.T("cmd.dev.vm.up_to_date")))
		return core.Ok(nil)
	}

	cli.Text(warningStyle.Render(i18n.T("cmd.dev.vm.update_available")))
	cli.Blank()

	if !apply {
		cli.Text(i18n.T("cmd.dev.vm.run_to_update", map[string]any{"Command": dimStyle.Render("core dev update --apply")}))
		return core.Ok(nil)
	}

	// Stop if running
	running, r := d.IsRunning(ctx)
	if !r.OK {
		return cli.Wrap(r.Value.(error), "failed to check VM state")
	}
	if running {
		cli.Text(i18n.T("cmd.dev.vm.stopping_current"))
		if r := d.Stop(ctx); !r.OK {
			return cli.Wrap(r.Value.(error), "failed to stop current VM")
		}
	}

	cli.Text(i18n.T("cmd.dev.vm.downloading_update"))
	cli.Blank()

	start := time.Now()
	r = d.Install(ctx, func(downloaded, total int64) {
		if total > 0 {
			pct := int(float64(downloaded) / float64(total) * 100)
			cli.Print("\r%s %d%%", dimStyle.Render(i18n.T("cmd.dev.vm.progress_label")), pct)
		}
	})

	cli.Blank()

	if !r.OK {
		return cli.Wrap(r.Value.(error), "update failed")
	}

	elapsed := time.Since(start).Round(time.Second)
	cli.Blank()
	cli.Text(i18n.T("cmd.dev.vm.updated_in", map[string]any{"Duration": elapsed}))

	return core.Ok(nil)
}
