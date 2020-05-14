// CLI entry point for gogit binary.
package main

import (
	"flag"
	"fmt"
	"os"
)

// Interface that all subcommands must implement.
type Subcommand interface {
	Init([]string) error
	Name() string
	Description() string
	Validate() error
	Usage()
	Execute()
}

// Helper function to exit on irrecoverable error.
func Check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Parse CLI arguments and execute the subcommand.
func execute(progName string, args []string) {
	// Create an object for each subcommand.
	cmds := []Subcommand{
		NewInitCommand(),
		NewHashObjectCommand(),
		NewCatFileCommand(),
		NewLsTreeCommand(),
		NewMkTreeCommand(),
		NewCheckoutCommand(),
	}

	// Prepare the global usage message.
	flag.Usage = func() {
		fmt.Println("gogit - the stupid content tracker\n")
		fmt.Printf("usage: %s <command> [<args>]\n", progName)
		fmt.Println("Valid commands:")

		for _, cmd := range cmds {
			fmt.Printf("  %-14s %s\n", cmd.Name(), cmd.Description())
		}
		flag.PrintDefaults()
	}

	if len(args) < 1 {
		flag.Usage()
		return
	}

	subcommand := os.Args[1]
	for _, cmd := range cmds {
		if cmd.Name() != subcommand {
			continue
		}

		// Parse the optional arguments.
		cmd.Init(os.Args[2:])

		// Validate the command specific arguments.
		if err := cmd.Validate(); err != nil {
			fmt.Println(err)
			cmd.Usage()
			os.Exit(1)
		}

		// Execute this command.
		cmd.Execute()
		return
	}

	fmt.Errorf("%[1]s: '%s' is not a valid command. See '%[1]s --help'",
		progName, subcommand)
	flag.Usage()
}

func main() {
	execute(os.Args[0], os.Args[1:])
}
