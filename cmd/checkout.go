package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

// CheckoutCommand lists the components of "checkout" comamnd.
type CheckoutCommand struct {
	fs       *flag.FlagSet
	path     string
	revision string
}

// NewCheckoutCommand creates a new command object.
func NewCheckoutCommand() *CheckoutCommand {
	cmd := &CheckoutCommand{
		fs: flag.NewFlagSet("checkout", flag.ExitOnError),
	}

	cmd.fs.StringVar(&cmd.path, "path", ".", "Path to create the files")
	return cmd
}

// Name gives the name of the command.
func (cmd *CheckoutCommand) Name() string {
	return cmd.fs.Name()
}

// Description gives the description of the command.
func (cmd *CheckoutCommand) Description() string {
	return "restore working tree files"
}

// Init initializes and validates the given command.
func (cmd *CheckoutCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		return errors.New("Error: Missing <object> argument\n")
	}

	cmd.revision = cmd.fs.Arg(0)
	return nil
}

// Usage prints the usage string for the end user.
func (cmd *CheckoutCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] <object>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *CheckoutCommand) Execute() {
	repo, err := git.GetRepo(".")
	util.Check(err)

	// Resolve the given hash to a full hash.
	objHash, err := repo.ObjectFind(cmd.revision)
	util.Check(err)

	obj, err := repo.ObjectParse(objHash)
	if err != nil {
		fmt.Println("fatal: not a tree object.", err)
		os.Exit(1)
	}
	if obj.ObjType != "tree" && obj.ObjType != "commit" {
		fmt.Println("fatal: not a tree object")
		os.Exit(1)
	}

	// If it is a "commit" object, then get its "tree" component first.
	if obj.ObjType == "commit" {
		commit, err := git.NewCommit(obj)
		util.Check(err)
		obj, err = repo.ObjectParse(commit.TreeHash())
		util.Check(err)
	}

	// "obj" is now a valid tree object.
	tree, err := git.NewTree(obj)
	util.Check(err)

	err = tree.Checkout(cmd.path)
	util.Check(err)
}
