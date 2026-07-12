package main

import (
	"fmt"
	"os"

	"github.com/ChaosHour/ssh-audits/internal/cli"
)

// version is set at build time via -ldflags "-X main.version=v1.2.3".
var version = "dev"

func main() {
	if err := cli.Run(version); err != nil {
		fmt.Fprintln(os.Stderr, "ssh-audits:", err)
		os.Exit(1)
	}
}
