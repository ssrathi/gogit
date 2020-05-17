package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ssrathi/gogit/git"
)

type MkTreeCommand struct {
	fs *flag.FlagSet
}

func NewMkTreeCommand() *MkTreeCommand {
	cmd := &MkTreeCommand{
		fs: flag.NewFlagSet("mktree", flag.ExitOnError),
	}
	return cmd
}

func (cmd *MkTreeCommand) Name() string {
	return cmd.fs.Name()
}

func (cmd *MkTreeCommand) Description() string {
	return "Build a tree-object from ls-tree formatted text"
}

func (cmd *MkTreeCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	return cmd.fs.Parse(args)
}

func (cmd *MkTreeCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

func (cmd *MkTreeCommand) Execute() {
	repo, err := git.GetRepo(".")
	Check(err)

	input, err := ioutil.ReadAll(os.Stdin)
	Check(err)

	tree, err := git.NewTreeFromInput(string(input))
	Check(err)

	// Write the tree now.
	hash, err := repo.ObjectWrite(tree.Obj, true)
	Check(err)

	fmt.Println(hash)
}
