package devkit

import core "dappco.re/go"

func ExampleScanSecrets() {
	old := scanSecretsRunner
	scanSecretsRunner = func(string) ([]byte, core.Result) {
		output := []byte("RuleID,File,StartLine,StartColumn,Match\ngithub-token,config.yml,2,3,ghp_exampletoken1234567890\n")
		return output, core.Ok(output)
	}
	defer func() { scanSecretsRunner = old }()
	findings, r := ScanSecrets("/tmp/project")
	core.Println(r.OK, findings[0].Rule)
	// Output: true github-token
}
