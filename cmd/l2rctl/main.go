package main

import (
	"fmt"
	"os"
)

const usage = `Usage: l2rctl <command> [options]

Commands:
  start   Start l2radar containers (probe, ui, or all)
  stop    Stop l2radar containers
  status  Show container status
  dump    Dump neighbour table

Use "l2rctl <command> --help" for more information about a command.`

var validSubcommands = map[string]bool{
	"start":  true,
	"stop":   true,
	"status": true,
	"dump":   true,
}

func parseSubcommand(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("no subcommand specified")
	}
	sub := args[1]
	if sub == "--help" || sub == "-h" {
		return "help", nil
	}
	if !validSubcommands[sub] {
		return "", fmt.Errorf("unknown subcommand: %s", sub)
	}
	return sub, nil
}

func run(args []string) error {
	sub, err := parseSubcommand(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usage)
		return err
	}
	if sub == "help" {
		fmt.Println(usage)
		return nil
	}

	switch sub {
	case "start":
		fmt.Fprintln(os.Stderr, "start: not yet implemented")
	case "stop":
		fmt.Fprintln(os.Stderr, "stop: not yet implemented")
	case "status":
		fmt.Fprintln(os.Stderr, "status: not yet implemented")
	case "dump":
		fmt.Fprintln(os.Stderr, "dump: not yet implemented")
	}
	return nil
}

func main() {
	if err := run(os.Args); err != nil {
		os.Exit(1)
	}
}
