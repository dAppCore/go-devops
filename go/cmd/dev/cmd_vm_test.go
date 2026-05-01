package dev

import (
	core "dappco.re/go"
	"slices"
	"testing"

	"dappco.re/go/cli/pkg/cli"
)

func newCmdVMTestDevEnv(t *testing.T, installed bool) *DevEnv {
	t.Helper()
	imageDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", imageDir)
	if installed {
		if r := core.WriteFile(core.PathJoin(imageDir, vmImageName()), []byte("image"), 0o600); !r.OK {
			t.Fatalf("write VM image: %v", r.Value)
		}
	}
	env, r := newVMDevEnv()
	if !r.OK {
		t.Fatalf("create VM environment: %v", r.Value)
	}
	return env
}

func requireVMResultFailure(t *testing.T, operation string, r core.Result) string {
	t.Helper()
	if r.OK {
		t.Fatalf("%s result OK = true, want false", operation)
	}
	err, ok := r.Value.(error)
	if !ok {
		t.Fatalf("%s result value = %T, want error", operation, r.Value)
	}
	message := err.Error()
	if operation != "" && !core.Contains(message, operation) {
		t.Fatalf("%s error = %q, want operation name", operation, message)
	}
	return message
}

func TestAddVMStatusCommand_Good(t *testing.T) {
	root := &cli.Command{Use: "core"}

	AddDevCommands(root)

	statusCmd, _, err := root.Find([]string{"dev", "status"})
	if err != nil {
		t.Fatalf("find status command: %v", err)
	}
	if statusCmd == nil {
		t.Fatal("expected status command")
	}
	if statusCmd.Use != "status" {
		t.Fatalf("status command use = %q, want %q", statusCmd.Use, "status")
	}
	if !slices.Contains(statusCmd.Aliases, "vm-status") {
		t.Fatalf("status aliases = %v, want vm-status", statusCmd.Aliases)
	}

	aliasCmd, _, err := root.Find([]string{"dev", "vm-status"})
	if err != nil {
		t.Fatalf("find vm-status alias: %v", err)
	}
	if aliasCmd == nil {
		t.Fatal("expected vm-status alias command")
	}
	if statusCmd != aliasCmd {
		t.Fatal("expected vm-status alias to resolve to status command")
	}
}

func TestCmdVm_DevEnv_IsInstalled_Good(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	installed := env.IsInstalled()
	if !installed {
		t.Fatal("IsInstalled() = false, want true")
	}
}

func TestCmdVm_DevEnv_IsInstalled_Bad(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	installed := env.IsInstalled()
	if installed {
		t.Fatal("IsInstalled() = true, want false")
	}
}

func TestCmdVm_DevEnv_IsInstalled_Ugly(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	if r := core.Remove(core.PathJoin(core.Getenv("CORE_IMAGES_DIR"), vmImageName())); !r.OK {
		t.Fatalf("remove VM image: %v", r.Value)
	}
	if env.IsInstalled() {
		t.Fatal("IsInstalled() = true after image removal, want false")
	}
}

func TestCmdVm_DevEnv_Install_Good(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	progressCalled := false
	r := env.Install(core.Background(), func(downloaded, total int64) {
		progressCalled = downloaded > 0 || total > 0
	})
	requireVMResultFailure(t, "installer", r)
	if progressCalled {
		t.Fatal("Install called progress callback for unavailable backend")
	}
}

func TestCmdVm_DevEnv_Install_Bad(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	r := env.Install(core.Background(), nil)
	requireVMResultFailure(t, "installer", r)
}

func TestCmdVm_DevEnv_Install_Ugly(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	ctx, cancel := core.WithCancel(core.Background())
	cancel()
	r := env.Install(ctx, func(downloaded, total int64) {})
	requireVMResultFailure(t, "installer", r)
}

func TestCmdVm_DevEnv_Boot_Good(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	r := env.Boot(core.Background(), vmBootOptions{Memory: 2048, CPUs: 1, Name: "test-vm"})
	requireVMResultFailure(t, "boot", r)
}

func TestCmdVm_DevEnv_Boot_Bad(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	r := env.Boot(core.Background(), defaultVMBootOptions())
	if r.OK {
		t.Fatal("Boot() OK = true for missing image, want false")
	}
}

func TestCmdVm_DevEnv_Boot_Ugly(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	r := env.Boot(core.Background(), vmBootOptions{Memory: 8192, CPUs: 8, Name: "fresh", Fresh: true})
	requireVMResultFailure(t, "boot", r)
}

func TestCmdVm_DevEnv_Stop_Good(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	r := env.Stop(core.Background())
	requireVMResultFailure(t, "stop", r)
}

func TestCmdVm_DevEnv_Stop_Bad(t *testing.T) {
	var env *DevEnv
	r := env.Stop(core.Background())
	requireVMResultFailure(t, "stop", r)
}

func TestCmdVm_DevEnv_Stop_Ugly(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	ctx, cancel := core.WithCancel(core.Background())
	cancel()
	r := env.Stop(ctx)
	requireVMResultFailure(t, "stop", r)
}

func TestCmdVm_DevEnv_IsRunning_Good(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	running, r := env.IsRunning(core.Background())
	if !r.OK || running {
		t.Fatalf("IsRunning() = (%v, %v), want (false, OK)", running, r.OK)
	}
}

func TestCmdVm_DevEnv_IsRunning_Bad(t *testing.T) {
	var env *DevEnv
	running, r := env.IsRunning(core.Background())
	if !r.OK || running {
		t.Fatalf("IsRunning() nil receiver = (%v, %v), want (false, OK)", running, r.OK)
	}
}

func TestCmdVm_DevEnv_IsRunning_Ugly(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	ctx, cancel := core.WithCancel(core.Background())
	cancel()
	running, r := env.IsRunning(ctx)
	if !r.OK || running {
		t.Fatalf("IsRunning() canceled = (%v, %v), want (false, OK)", running, r.OK)
	}
}

func TestCmdVm_DevEnv_Status_Good(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	status, r := env.Status(core.Background())
	if !r.OK || !status.Installed || status.SSHPort != vmDefaultSSHPort {
		t.Fatalf("Status() = (%+v, %v), want installed with default SSH port", status, r.OK)
	}
}

func TestCmdVm_DevEnv_Status_Bad(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	status, r := env.Status(core.Background())
	if !r.OK || status.Installed {
		t.Fatalf("Status() = (%+v, %v), want not installed with OK result", status, r.OK)
	}
}

func TestCmdVm_DevEnv_Status_Ugly(t *testing.T) {
	var env *DevEnv
	t.Setenv("CORE_IMAGES_DIR", t.TempDir())
	status, r := env.Status(core.Background())
	if !r.OK || status.Running || status.Memory != 4096 || status.CPUs != 2 {
		t.Fatalf("Status() nil receiver = (%+v, %v), want stopped defaults", status, r.OK)
	}
}

func TestCmdVm_DevEnv_Shell_Good(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	r := env.Shell(core.Background(), vmShellOptions{Command: []string{"echo", "ok"}})
	requireVMResultFailure(t, "shell", r)
}

func TestCmdVm_DevEnv_Shell_Bad(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	r := env.Shell(core.Background(), vmShellOptions{})
	requireVMResultFailure(t, "shell", r)
}

func TestCmdVm_DevEnv_Shell_Ugly(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	r := env.Shell(core.Background(), vmShellOptions{Console: true})
	requireVMResultFailure(t, "shell", r)
}

func TestCmdVm_DevEnv_Serve_Good(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	r := env.Serve(core.Background(), t.TempDir(), vmServeOptions{Port: 3000})
	requireVMResultFailure(t, "serve", r)
}

func TestCmdVm_DevEnv_Serve_Bad(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	r := env.Serve(core.Background(), "", vmServeOptions{})
	requireVMResultFailure(t, "serve", r)
}

func TestCmdVm_DevEnv_Serve_Ugly(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	ctx, cancel := core.WithCancel(core.Background())
	cancel()
	r := env.Serve(ctx, t.TempDir(), vmServeOptions{Port: -1, Path: "../outside"})
	requireVMResultFailure(t, "serve", r)
}

func TestCmdVm_DevEnv_Test_Good(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	r := env.Test(core.Background(), t.TempDir(), vmTestOptions{Name: "unit", Command: []string{"go", "test", "./..."}})
	requireVMResultFailure(t, "test", r)
}

func TestCmdVm_DevEnv_Test_Bad(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	r := env.Test(core.Background(), "", vmTestOptions{})
	requireVMResultFailure(t, "test", r)
}

func TestCmdVm_DevEnv_Test_Ugly(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	ctx, cancel := core.WithCancel(core.Background())
	cancel()
	r := env.Test(ctx, t.TempDir(), vmTestOptions{Name: "canceled"})
	requireVMResultFailure(t, "test", r)
}

func TestCmdVm_DevEnv_Claude_Good(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	r := env.Claude(core.Background(), t.TempDir(), vmClaudeOptions{Model: "sonnet"})
	requireVMResultFailure(t, "Claude", r)
}

func TestCmdVm_DevEnv_Claude_Bad(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	r := env.Claude(core.Background(), "", vmClaudeOptions{NoAuth: true})
	requireVMResultFailure(t, "Claude", r)
}

func TestCmdVm_DevEnv_Claude_Ugly(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	ctx, cancel := core.WithCancel(core.Background())
	cancel()
	r := env.Claude(ctx, t.TempDir(), vmClaudeOptions{Auth: []string{"ANTHROPIC_API_KEY"}})
	requireVMResultFailure(t, "Claude", r)
}

func TestCmdVm_DevEnv_CheckUpdate_Good(t *testing.T) {
	env := newCmdVMTestDevEnv(t, false)
	current, latest, hasUpdate, r := env.CheckUpdate(core.Background())
	if !r.OK || current != "unknown" || latest != "unknown" || hasUpdate {
		t.Fatalf("CheckUpdate() = (%q, %q, %v, %v), want unknown unknown false OK", current, latest, hasUpdate, r.OK)
	}
}

func TestCmdVm_DevEnv_CheckUpdate_Bad(t *testing.T) {
	var env *DevEnv
	current, latest, hasUpdate, r := env.CheckUpdate(core.Background())
	if !r.OK || current != "unknown" || latest != "unknown" || hasUpdate {
		t.Fatalf("CheckUpdate() nil receiver = (%q, %q, %v, %v), want unknown unknown false OK", current, latest, hasUpdate, r.OK)
	}
}

func TestCmdVm_DevEnv_CheckUpdate_Ugly(t *testing.T) {
	env := newCmdVMTestDevEnv(t, true)
	ctx, cancel := core.WithCancel(core.Background())
	cancel()
	current, latest, hasUpdate, r := env.CheckUpdate(ctx)
	if !r.OK || current != "unknown" || latest != "unknown" || hasUpdate {
		t.Fatalf("CheckUpdate() canceled = (%q, %q, %v, %v), want unknown unknown false OK", current, latest, hasUpdate, r.OK)
	}
}
