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
		if dir != "/tmp/project" {
			t.Fatalf("dir = %q, want /tmp/project", dir)
		}
		return []byte(`RuleID,File,StartLine,StartColumn,Description,Match
github-token,config.yml,12,4,GitHub token detected,ghp_exampletoken1234567890
aws-access-key-id,creds.txt,7,1,AWS access key detected,AKIA1234567890ABCDEF
`), nil
	}

	findings, err := ScanSecrets("/tmp/project")
	if err != nil {
		t.Fatalf("scan secrets: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("findings length = %d, want 2", len(findings))
	}

	if findings[0].Rule != "github-token" {
		t.Fatalf("findings[0].Rule = %q, want github-token", findings[0].Rule)
	}
	if findings[0].Path != "config.yml" {
		t.Fatalf("findings[0].Path = %q, want config.yml", findings[0].Path)
	}
	if findings[0].Line != 12 {
		t.Fatalf("findings[0].Line = %d, want 12", findings[0].Line)
	}
	if findings[0].Column != 4 {
		t.Fatalf("findings[0].Column = %d, want 4", findings[0].Column)
	}
	if findings[0].Snippet != "ghp_exampletoken1234567890" {
		t.Fatalf("findings[0].Snippet = %q, want ghp_exampletoken1234567890", findings[0].Snippet)
	}

	if findings[1].Rule != "aws-access-key-id" {
		t.Fatalf("findings[1].Rule = %q, want aws-access-key-id", findings[1].Rule)
	}
	if findings[1].Path != "creds.txt" {
		t.Fatalf("findings[1].Path = %q, want creds.txt", findings[1].Path)
	}
	if findings[1].Line != 7 {
		t.Fatalf("findings[1].Line = %d, want 7", findings[1].Line)
	}
	if findings[1].Column != 1 {
		t.Fatalf("findings[1].Column = %d, want 1", findings[1].Column)
	}
	if findings[1].Snippet != "AKIA1234567890ABCDEF" {
		t.Fatalf("findings[1].Snippet = %q, want AKIA1234567890ABCDEF", findings[1].Snippet)
	}
}

func TestScanSecrets_ReportsFindingsOnExitError_Good(t *testing.T) {
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
	if err != nil {
		t.Fatalf("scan secrets: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings length = %d, want 1", len(findings))
	}
	if findings[0].Rule != "token" {
		t.Fatalf("findings[0].Rule = %q, want token", findings[0].Rule)
	}
	if findings[0].Line != 3 {
		t.Fatalf("findings[0].Line = %d, want 3", findings[0].Line)
	}
	if findings[0].Column != 2 {
		t.Fatalf("findings[0].Column = %d, want 2", findings[0].Column)
	}
}

func TestParseGitleaksCSV_Bad(t *testing.T) {
	_, err := parseGitleaksCSV([]byte("rule_id,file,start_line\nunterminated,\"broken"))
	if err == nil {
		t.Fatal("expected parse error")
	}
}
