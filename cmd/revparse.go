package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

// RevParseCommand lists the components of "rev-parse" comamnd.
type RevParseCommand struct {
	fs       *flag.FlagSet
	revision string
}

// NewRevParseCommand creates a new command object.
func NewRevParseCommand() *RevParseCommand {
	cmd := &RevParseCommand{
		fs: flag.NewFlagSet("rev-parse", flag.ExitOnError),
	}
	return cmd
}

// Name gives the name of the command.
func (cmd *RevParseCommand) Name() string {
	return cmd.fs.Name()
}

// Init initializes and validates the given command.
func (cmd *RevParseCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		return errors.New("error: Missing <identifier> argument")
	}

	cmd.revision = cmd.fs.Arg(0)
	return nil
}

// Description gives the description of the command.
func (cmd *RevParseCommand) Description() string {
	return "Parse a given git identifier"
}

// Usage prints the usage string for the end user.
func (cmd *RevParseCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s <identifier>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *RevParseCommand) Execute() {
	repo, err := git.GetRepo(".")
	util.Check(err)

	// Resolve the given name to a full hash.
	objHash, err := repo.UniqueNameResolve(cmd.revision)
	util.Check(err)

	fmt.Println(objHash)
}
