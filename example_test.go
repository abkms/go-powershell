package powershell_test

import (
	"fmt"
	"log"

	"github.com/ycyun/go-powershell"
)

func Example() {
	shell, err := powershell.New()
	if err != nil {
		log.Fatal(err)
	}
	defer shell.Exit()

	stdout, err := shell.Exec("echo こんにちは")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stdout)

	stdout, err = shell.Exec("Get-TimeZone | Select-Object StandardName")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stdout)
}
