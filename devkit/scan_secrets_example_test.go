package devkit

import . "dappco.re/go"

func ExampleScanSecrets() {
	old := scanSecretsRunner
	scanSecretsRunner = func(string) ([]byte, error) {
		return []byte("RuleID,File,StartLine,StartColumn,Match\ngithub-token,config.yml,2,3,ghp_exampletoken1234567890\n"), nil
	}
	defer func() { scanSecretsRunner = old }()
	findings, err := ScanSecrets("/tmp/project")
	Println(err == nil, findings[0].Rule)
	// Output: true github-token
}
