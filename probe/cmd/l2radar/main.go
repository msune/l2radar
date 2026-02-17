package main

import (
	"os"

	"github.com/marc/l2radar/probe/cmd/l2radar/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
