package setup

import . "dappco.re/go"

func securityExampleFakeGH(body string) func() {
	dir := MustCast[string](MkdirTemp("", "security-gh-*"))
	WriteFile(PathJoin(dir, "gh"), []byte("#!/bin/sh\n"+body+"\n"), 0o755)
	oldPath := Getenv("PATH")
	Setenv("PATH", dir+":"+oldPath)
	return func() { Setenv("PATH", oldPath); RemoveAll(dir) }
}

func securityExampleHappyGH() func() {
	return securityExampleFakeGH(`
if [ "$2" = "repos/owner/repo/vulnerability-alerts" ]; then exit 0; fi
if [ "$2" = "repos/owner/repo/automated-security-fixes" ]; then exit 0; fi
echo '{"security_and_analysis":{"secret_scanning":{"status":"disabled"},"secret_scanning_push_protection":{"status":"disabled"},"dependabot_security_updates":{"status":"disabled"}}}'
`)
}

func ExampleGetSecuritySettings() {
	cleanup := securityExampleHappyGH()
	defer cleanup()
	status, err := GetSecuritySettings("owner/repo")
	Println(err == nil, status.DependabotAlerts)
	// Output: true true
}

func ExampleEnableDependabotAlerts() {
	cleanup := securityExampleHappyGH()
	defer cleanup()
	err := EnableDependabotAlerts("owner/repo")
	Println(err == nil)
	// Output: true
}

func ExampleEnableDependabotSecurityUpdates() {
	cleanup := securityExampleHappyGH()
	defer cleanup()
	err := EnableDependabotSecurityUpdates("owner/repo")
	Println(err == nil)
	// Output: true
}

func ExampleDisableDependabotSecurityUpdates() {
	cleanup := securityExampleHappyGH()
	defer cleanup()
	err := DisableDependabotSecurityUpdates("owner/repo")
	Println(err == nil)
	// Output: true
}

func ExampleUpdateSecurityAndAnalysis() {
	cleanup := securityExampleHappyGH()
	defer cleanup()
	err := UpdateSecurityAndAnalysis("owner/repo", true, true)
	Println(err == nil)
	// Output: true
}

func ExampleSyncSecuritySettings() {
	cleanup := securityExampleHappyGH()
	defer cleanup()
	changes, err := SyncSecuritySettings("owner/repo", &GitHubConfig{Security: SecurityConfig{SecretScanning: true}}, true)
	Println(err == nil, changes.HasChanges())
	// Output: true true
}
