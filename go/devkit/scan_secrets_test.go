package devkit

import (
	core "dappco.re/go"
)

func TestScanSecrets_ScanSecrets_Good(t *core.T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, core.Result) {
		output := []byte("RuleID,File,StartLine,StartColumn,Match\ngithub-token,config.yml,2,3,ghp_exampletoken1234567890\n")
		return output, core.Ok(output)
	}

	findings, r := ScanSecrets("/tmp/project")
	core.AssertTrue(t, r.OK)
	core.AssertEqual(t, "github-token", findings[0].Rule)
}

func TestScanSecrets_ScanSecrets_Bad(t *core.T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, core.Result) { return nil, core.Fail(core.AnError) }

	findings, r := ScanSecrets("/tmp/project")
	core.AssertFalse(t, r.OK)
	core.AssertNil(t, findings)
}

func TestScanSecrets_ScanSecrets_Ugly(t *core.T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, core.Result) { return nil, core.Ok(nil) }

	findings, r := ScanSecrets("/tmp/project")
	core.AssertTrue(t, r.OK)
	core.AssertNil(t, findings)
}
