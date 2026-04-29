package devkit

import (
	. "dappco.re/go"
)

func TestScanSecrets_ScanSecrets_Good(t *T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, error) {
		return []byte("RuleID,File,StartLine,StartColumn,Match\ngithub-token,config.yml,2,3,ghp_exampletoken1234567890\n"), nil
	}

	findings, err := ScanSecrets("/tmp/project")
	AssertNoError(t, err)
	AssertEqual(t, "github-token", findings[0].Rule)
}

func TestScanSecrets_ScanSecrets_Bad(t *T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, error) { return nil, AnError }

	findings, err := ScanSecrets("/tmp/project")
	AssertError(t, err)
	AssertNil(t, findings)
}

func TestScanSecrets_ScanSecrets_Ugly(t *T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, error) { return nil, nil }

	findings, err := ScanSecrets("/tmp/project")
	AssertNoError(t, err)
	AssertNil(t, findings)
}
