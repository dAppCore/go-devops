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

// addVMCommands adds the dev environment VM commands to the dev parent command.
// These are added as direct subcommands: core dev install, core dev boot, etc.
func addVMCommands(parent *cli.Command) {
	addVMInstallCommand(parent)
	addVMBootCommand(parent)
	addVMStopCommand(parent)
	addVMStatusCommand(parent)
	addVMShellCommand(parent)
	addVMServeCommand(parent)
	addVMTestCommand(parent)
	addVMClaudeCommand(parent)
	addVMUpdateCommand(parent)
}

// addVMInstallCommand adds the 'dev install' command.
func addVMInstallCommand(parent *cli.Command) {
	installCmd := &cli.Command{
		Use:   "install",
		Short: i18n.T("cmd.dev.vm.install.short"),
		Long:  i18n.T("cmd.dev.vm.install.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runVMInstall())
		},
	}

	parent.AddCommand(installCmd)
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
		return core.Fail(cli.Wrap(r.Value.(error), "install failed"))
	}

	elapsed := time.Since(start).Round(time.Second)
	cli.Blank()
	cli.Text(i18n.T("cmd.dev.vm.installed_in", map[string]any{"Duration": elapsed}))
	cli.Blank()
	cli.Text(i18n.T("cmd.dev.vm.start_with", map[string]any{"Command": dimStyle.Render("core dev boot")}))

	return core.Ok(nil)
}

// VM boot command flags
var (
	vmBootMemory int
	vmBootCPUs   int
	vmBootFresh  bool
)

// addVMBootCommand adds the 'devops boot' command.
func addVMBootCommand(parent *cli.Command) {
	bootCmd := &cli.Command{
		Use:   "boot",
		Short: i18n.T("cmd.dev.vm.boot.short"),
		Long:  i18n.T("cmd.dev.vm.boot.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runVMBoot(vmBootMemory, vmBootCPUs, vmBootFresh))
		},
	}

	bootCmd.Flags().IntVar(&vmBootMemory, "memory", 0, i18n.T("cmd.dev.vm.boot.flag.memory"))
	bootCmd.Flags().IntVar(&vmBootCPUs, "cpus", 0, i18n.T("cmd.dev.vm.boot.flag.cpus"))
	bootCmd.Flags().BoolVar(&vmBootFresh, "fresh", false, i18n.T("cmd.dev.vm.boot.flag.fresh"))

	parent.AddCommand(bootCmd)
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

// addVMStopCommand adds the 'devops stop' command.
func addVMStopCommand(parent *cli.Command) {
	stopCmd := &cli.Command{
		Use:   "stop",
		Short: i18n.T("cmd.dev.vm.stop.short"),
		Long:  i18n.T("cmd.dev.vm.stop.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runVMStop())
		},
	}

	parent.AddCommand(stopCmd)
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

// addVMStatusCommand adds the 'dev status' command.
func addVMStatusCommand(parent *cli.Command) {
	statusCmd := &cli.Command{
		Use: "status",
		Aliases: []string{
			"vm-status",
		},
		Short: i18n.T("cmd.dev.vm.status.short"),
		Long:  i18n.T("cmd.dev.vm.status.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runVMStatus())
		},
	}

	parent.AddCommand(statusCmd)
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

// VM shell command flags
var vmShellConsole bool

// addVMShellCommand adds the 'devops shell' command.
func addVMShellCommand(parent *cli.Command) {
	shellCmd := &cli.Command{
		Use:   "shell [-- command...]",
		Short: i18n.T("cmd.dev.vm.shell.short"),
		Long:  i18n.T("cmd.dev.vm.shell.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runVMShell(vmShellConsole, args))
		},
	}

	shellCmd.Flags().BoolVar(&vmShellConsole, "console", false, i18n.T("cmd.dev.vm.shell.flag.console"))

	parent.AddCommand(shellCmd)
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

// VM serve command flags
var (
	vmServePort int
	vmServePath string
)

// addVMServeCommand adds the 'devops serve' command.
func addVMServeCommand(parent *cli.Command) {
	serveCmd := &cli.Command{
		Use:   "serve",
		Short: i18n.T("cmd.dev.vm.serve.short"),
		Long:  i18n.T("cmd.dev.vm.serve.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runVMServe(vmServePort, vmServePath))
		},
	}

	serveCmd.Flags().IntVarP(&vmServePort, "port", "p", 0, i18n.T("cmd.dev.vm.serve.flag.port"))
	serveCmd.Flags().StringVar(&vmServePath, "p"+"ath", "", i18n.T("cmd.dev.vm.serve.flag.path"))

	parent.AddCommand(serveCmd)
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

// VM test command flags
var vmTestName string

// addVMTestCommand adds the 'devops test' command.
func addVMTestCommand(parent *cli.Command) {
	testCmd := &cli.Command{
		Use:   "test [-- command...]",
		Short: i18n.T("cmd.dev.vm.test.short"),
		Long:  i18n.T("cmd.dev.vm.test.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runVMTest(vmTestName, args))
		},
	}

	testCmd.Flags().StringVarP(&vmTestName, "name", "n", "", i18n.T("cmd.dev.vm.test.flag.name"))

	parent.AddCommand(testCmd)
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

// VM claude command flags
var (
	vmClaudeNoAuth    bool
	vmClaudeModel     string
	vmClaudeAuthFlags []string
)

// addVMClaudeCommand adds the 'devops claude' command.
func addVMClaudeCommand(parent *cli.Command) {
	claudeCmd := &cli.Command{
		Use:   "claude",
		Short: i18n.T("cmd.dev.vm.claude.short"),
		Long:  i18n.T("cmd.dev.vm.claude.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runVMClaude(vmClaudeNoAuth, vmClaudeModel, vmClaudeAuthFlags))
		},
	}

	claudeCmd.Flags().BoolVar(&vmClaudeNoAuth, "no-auth", false, i18n.T("cmd.dev.vm.claude.flag.no_auth"))
	claudeCmd.Flags().StringVarP(&vmClaudeModel, "model", "m", "", i18n.T("cmd.dev.vm.claude.flag.model"))
	claudeCmd.Flags().StringSliceVar(&vmClaudeAuthFlags, "auth", nil, i18n.T("cmd.dev.vm.claude.flag.auth"))

	parent.AddCommand(claudeCmd)
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

// VM update command flags
var vmUpdateApply bool

// addVMUpdateCommand adds the 'devops update' command.
func addVMUpdateCommand(parent *cli.Command) {
	updateCmd := &cli.Command{
		Use:   "update",
		Short: i18n.T("cmd.dev.vm.update.short"),
		Long:  i18n.T("cmd.dev.vm.update.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runVMUpdate(vmUpdateApply))
		},
	}

	updateCmd.Flags().BoolVar(&vmUpdateApply, "apply", false, i18n.T("cmd.dev.vm.update.flag.apply"))

	parent.AddCommand(updateCmd)
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
		return core.Fail(cli.Wrap(r.Value.(error), "failed to check for updates"))
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
		return core.Fail(cli.Wrap(r.Value.(error), "failed to check VM state"))
	}
	if running {
		cli.Text(i18n.T("cmd.dev.vm.stopping_current"))
		if r := d.Stop(ctx); !r.OK {
			return core.Fail(cli.Wrap(r.Value.(error), "failed to stop current VM"))
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
		return core.Fail(cli.Wrap(r.Value.(error), "update failed"))
	}

	elapsed := time.Since(start).Round(time.Second)
	cli.Blank()
	cli.Text(i18n.T("cmd.dev.vm.updated_in", map[string]any{"Duration": elapsed}))

	return core.Ok(nil)
}
