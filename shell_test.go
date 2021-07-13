package powershell_test

import (
	"strings"
	"testing"

	"github.com/ycyun/go-powershell"
)

func TestShell_Exec(t *testing.T) {
	shell, err := powershell.New()
	if err != nil {
		t.Fatal(err)
	}
	defer shell.Exit()

	cases := []struct {
		command string
		want    string
	}{
		{"echo こんにちは", "こんにちは"},
		{"Get-TimeZone | Select-Object StandardName",
			func() string {
				if shell.CodePage() == 65001 {
					return "StandardName       \r\n------------       \r\nTokyo Standard Time"
				}
				return "StandardName \r\n------------ \r\n東京 (標準時)"
			}(),
		},
	}

	for i, c := range cases {
		stdout, err := shell.Exec(c.command)
		if err != nil {
			t.Errorf("error from exec, caseID=%d, commnad=%s: %s", i, c.command, err)
		}
		got := strings.TrimSpace(stdout)
		if got != c.want {
			t.Errorf("unexpected output: got=%q, want=%q", got, c.want)
		}
	}
}
