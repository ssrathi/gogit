package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

// MkTreeCommand lists the components of "mktree" comamnd.
type MkTreeCommand struct {
	fs *flag.FlagSet
}

// NewMkTreeCommand creates a new command object.
func NewMkTreeCommand() *MkTreeCommand {
	cmd := &MkTreeCommand{
		fs: flag.NewFlagSet("mktree", flag.ExitOnError),
	}
	return cmd
}

// Name gives the name of the command.
func (cmd *MkTreeCommand) Name() string {
	return cmd.fs.Name()
}

// Description gives the description of the command.
func (cmd *MkTreeCommand) Description() string {
	return "Build a tree-object from ls-tree formatted text"
}

// Init initializes and validates the given command.
func (cmd *MkTreeCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	return cmd.fs.Parse(args)
}

// Usage prints the usage string for the end user.
func (cmd *MkTreeCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *MkTreeCommand) Execute() {
	repo, err := git.GetRepo(".")
	util.Check(err)

	input, err := ioutil.ReadAll(os.Stdin)
	util.Check(err)

	tree, err := git.NewTreeFromInput(repo, string(input))
	util.Check(err)

	// Write the tree now.
	hash, err := repo.ObjectWrite(tree.Object, true)
	util.Check(err)

	fmt.Println(hash)
}
