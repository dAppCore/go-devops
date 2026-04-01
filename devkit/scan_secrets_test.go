package devkit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanSecrets_Good(t *testing.T) {
	originalRunner := scanSecretsRunner
	t.Cleanup(func() {
		scanSecretsRunner = originalRunner
	})

	scanSecretsRunner = func(dir string) ([]byte, error) {
		require.Equal(t, "/tmp/project", dir)
		return []byte(`RuleID,File,StartLine,StartColumn,Description,Match
github-token,config.yml,12,4,GitHub token detected,ghp_exampletoken1234567890
aws-access-key-id,creds.txt,7,1,AWS access key detected,AKIA1234567890ABCDEF
`), nil
	}

	findings, err := ScanSecrets("/tmp/project")
	require.NoError(t, err)
	require.Len(t, findings, 2)

	require.Equal(t, "github-token", findings[0].Rule)
	require.Equal(t, "config.yml", findings[0].Path)
	require.Equal(t, 12, findings[0].Line)
	require.Equal(t, 4, findings[0].Column)
	require.Equal(t, "ghp_exampletoken1234567890", findings[0].Snippet)

	require.Equal(t, "aws-access-key-id", findings[1].Rule)
	require.Equal(t, "creds.txt", findings[1].Path)
	require.Equal(t, 7, findings[1].Line)
	require.Equal(t, 1, findings[1].Column)
	require.Equal(t, "AKIA1234567890ABCDEF", findings[1].Snippet)
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
	require.NoError(t, err)
	require.Len(t, findings, 1)
	require.Equal(t, "token", findings[0].Rule)
	require.Equal(t, 3, findings[0].Line)
	require.Equal(t, 2, findings[0].Column)
}

func TestParseGitleaksCSV_Bad(t *testing.T) {
	_, err := parseGitleaksCSV([]byte("rule_id,file,start_line\nunterminated,\"broken"))
	require.Error(t, err)
}
