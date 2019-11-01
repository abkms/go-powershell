# go-powershell

go-powershell is a Go library to execute commands on a local PowerShell session.

This is a slimed down rewrite of github.com/bhendo/go-powershell
which is a fork of github.com/gorillalabs/go-powershell.

The API is not compatible with github.com/bhendo/go-powershell

The original package was inspired by [jPowerShell](https://github.com/profesorfalken/jPowerShell)
and allows one to run and remote-control a PowerShell session. Use this if you
don't have a static script that you want to execute, bur rather run dynamic
commands.

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/hnakamur/go-powershell"
)

func main() {
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
```

## License

MIT, see LICENSE file.
