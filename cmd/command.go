// Package cmd is the entry point for gogit command line parsing.
package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// Subcommand is an interface that all subcommands must implement.
type Subcommand interface {
	Init([]string) error
	Name() string
	Description() string
	Usage()
	Execute()
}

// Execute parses CLI arguments and executes the given subcommand.
func Execute() {
	progName := os.Args[0]
	args := os.Args[1:]

	// Create an object for each subcommand.
	cmds := []Subcommand{
		NewInitCommand(),
		NewHashObjectCommand(),
		NewCatFileCommand(),
		NewLsTreeCommand(),
		NewMkTreeCommand(),
		NewCheckoutCommand(),
		NewCommitTreeCommand(),
		NewLogCommand(),
		NewShowRefCommand(),
	}

	// Prepare the global usage message.
	flag.Usage = func() {
		fmt.Printf("gogit - the stupid content tracker\n\n")
		fmt.Printf("usage: %s <command> [<args>]\n", progName)
		fmt.Println("Valid commands:")

		for _, cmd := range cmds {
			fmt.Printf("  %-14s %s\n", cmd.Name(), cmd.Description())
		}
		flag.PrintDefaults()
		fmt.Printf("\nUse \"%s <command> --help\" for help on a specific "+
			"command\n", progName)
	}

	flag.Parse()
	if len(args) < 1 {
		flag.Usage()
		return
	}

	subcommand := args[0]
	for _, cmd := range cmds {
		if cmd.Name() != subcommand {
			continue
		}

		// Parse and validate the command specific arguments.
		if err := cmd.Init(args[1:]); err != nil {
			fmt.Println(err)
			fmt.Printf("See \"%s %s --help\".\n", progName, subcommand)
			os.Exit(1)
		}

		// Execute this command.
		log.Println("Executing command:", cmd.Name())
		cmd.Execute()
		return
	}

	fmt.Printf("%[1]s: '%s' is not a valid command. See '%[1]s --help'.\n",
		progName, subcommand)
}
