package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

// UpdateRefCommand lists the components of "update-ref" comamnd.
type UpdateRefCommand struct {
	fs        *flag.FlagSet
	reference string
	newValue  string
}

// NewUpdateRefCommand creates a new command object.
func NewUpdateRefCommand() *UpdateRefCommand {
	cmd := &UpdateRefCommand{
		fs: flag.NewFlagSet("update-ref", flag.ExitOnError),
	}

	return cmd
}

// Name gives the name of the command.
func (cmd *UpdateRefCommand) Name() string {
	return cmd.fs.Name()
}

// Init initializes and validates the given command.
func (cmd *UpdateRefCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 2 {
		return errors.New("error: <reference> and/or <new-value> not provided")
	}

	cmd.reference = cmd.fs.Arg(0)
	cmd.newValue = cmd.fs.Arg(1)
	return nil
}

// Description gives the description of the command.
func (cmd *UpdateRefCommand) Description() string {
	return "Update the object name stored in a ref safely"
}

// Usage prints the usage string for the end user.
func (cmd *UpdateRefCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] <reference> <new-value>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *UpdateRefCommand) Execute() {
	repo, err := git.GetRepo(".")
	util.Check(err)

	err = repo.UpdateRef(cmd.reference, cmd.newValue)
	util.Check(err)
}
