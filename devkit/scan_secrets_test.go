package devkit

import (
	. "dappco.re/go"
)

func TestScanSecrets_ScanSecrets_Good(t *T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, Result) {
		output := []byte("RuleID,File,StartLine,StartColumn,Match\ngithub-token,config.yml,2,3,ghp_exampletoken1234567890\n")
		return output, Ok(output)
	}

	findings, r := ScanSecrets("/tmp/project")
	AssertTrue(t, r.OK)
	AssertEqual(t, "github-token", findings[0].Rule)
}

func TestScanSecrets_ScanSecrets_Bad(t *T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, Result) { return nil, Fail(AnError) }

	findings, r := ScanSecrets("/tmp/project")
	AssertFalse(t, r.OK)
	AssertNil(t, findings)
}

func TestScanSecrets_ScanSecrets_Ugly(t *T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, Result) { return nil, Ok(nil) }

	findings, r := ScanSecrets("/tmp/project")
	AssertTrue(t, r.OK)
	AssertNil(t, findings)
}
