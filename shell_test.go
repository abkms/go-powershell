package powershell_test

import (
	"strings"
	"testing"

	"github.com/hnakamur/go-powershell"
)

func TestShell_Exec(t *testing.T) {
	t.Run("defaultCodePage", func(t *testing.T) {
		cases := []struct {
			command string
			want    string
		}{
			{"echo こんにちは", "こんにちは"},
			{"Get-TimeZone | Select-Object StandardName", "StandardName \r\n------------ \r\n東京 (標準時)"},
		}
		shell, err := powershell.New()
		if err != nil {
			t.Fatal(err)
		}
		defer shell.Exit()
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
	})
	t.Run("codePage932", func(t *testing.T) {
		cases := []struct {
			command string
			want    string
		}{
			{"echo こんにちは", "こんにちは"},
			{"Get-TimeZone | Select-Object StandardName", "StandardName \r\n------------ \r\n東京 (標準時)"},
		}
		shell, err := powershell.NewCodePage(932)
		if err != nil {
			t.Fatal(err)
		}
		defer shell.Exit()
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
	})
	t.Run("codePage65001", func(t *testing.T) {
		cases := []struct {
			command string
			want    string
		}{
			{"echo こんにちは", "こんにちは"},
			{"Get-TimeZone | Select-Object StandardName", "StandardName \r\n------------ \r\n東京 (標準時)"},
		}
		shell, err := powershell.NewCodePage(65001)
		if err != nil {
			t.Fatal(err)
		}
		defer shell.Exit()
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
	})
}
