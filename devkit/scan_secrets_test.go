package devkit

import (
	"errors"
	"testing"
)

func TestScanSecrets_Good(t *testing.T) {
	originalRunner := scanSecretsRunner
	t.Cleanup(func() {
		scanSecretsRunner = originalRunner
	})

	scanSecretsRunner = func(dir string) ([]byte, error) {
		mustEqual(t, "/tmp/project", dir)
		return []byte(`RuleID,File,StartLine,StartColumn,Description,Match
github-token,config.yml,12,4,GitHub token detected,ghp_exampletoken1234567890
aws-access-key-id,creds.txt,7,1,AWS access key detected,AKIA1234567890ABCDEF
`), nil
	}

	findings, err := ScanSecrets("/tmp/project")
	mustNoError(t, err)
	mustLen(t, findings, 2)

	mustEqual(t, "github-token", findings[0].Rule)
	mustEqual(t, "config.yml", findings[0].Path)
	mustEqual(t, 12, findings[0].Line)
	mustEqual(t, 4, findings[0].Column)
	mustEqual(t, "ghp_exampletoken1234567890", findings[0].Snippet)

	mustEqual(t, "aws-access-key-id", findings[1].Rule)
	mustEqual(t, "creds.txt", findings[1].Path)
	mustEqual(t, 7, findings[1].Line)
	mustEqual(t, 1, findings[1].Column)
	mustEqual(t, "AKIA1234567890ABCDEF", findings[1].Snippet)
}

func TestScanSecrets_ReportsFindingsOnExitError(t *testing.T) {
	originalRunner := scanSecretsRunner
	t.Cleanup(func() {
		scanSecretsRunner = originalRunner
	})

	scanSecretsRunner = func(dir string) ([]byte, error) {
		return []byte(`rule_id,file,start_line,start_column,description,match
token,test.txt,3,2,Token detected,secret-value
`), errors.New("exit status 1")
	}

	findings, err := ScanSecrets("/tmp/project")
	mustNoError(t, err)
	mustLen(t, findings, 1)
	mustEqual(t, "token", findings[0].Rule)
	mustEqual(t, 3, findings[0].Line)
	mustEqual(t, 2, findings[0].Column)
}

func TestParseGitleaksCSV_Bad(t *testing.T) {
	_, err := parseGitleaksCSV([]byte("rule_id,file,start_line\nunterminated,\"broken"))
	mustError(t, err)
}
