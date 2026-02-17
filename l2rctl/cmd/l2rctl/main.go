package main

import (
	"os"

	"github.com/msune/l2rctl/cmd/l2rctl/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
