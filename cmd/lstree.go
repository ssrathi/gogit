package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

// LsTreeCommand lists the components of "lstree" comamnd.
type LsTreeCommand struct {
	fs       *flag.FlagSet
	revision string
}

// NewLsTreeCommand creates a new command object.
func NewLsTreeCommand() *LsTreeCommand {
	cmd := &LsTreeCommand{
		fs: flag.NewFlagSet("ls-tree", flag.ExitOnError),
	}
	return cmd
}

// Name gives the name of the command.
func (cmd *LsTreeCommand) Name() string {
	return cmd.fs.Name()
}

// Init initializes and validates the given command.
func (cmd *LsTreeCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		return errors.New("Error: Missing <tree-ish> argument\n")
	}

	cmd.revision = cmd.fs.Arg(0)
	return nil
}

// Description gives the description of the command.
func (cmd *LsTreeCommand) Description() string {
	return "List the contents of a tree object"
}

// Usage prints the usage string for the end user.
func (cmd *LsTreeCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s <tree-ish>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *LsTreeCommand) Execute() {
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
		commit, err := git.NewCommit(repo, obj)
		util.Check(err)
		obj, err = repo.ObjectParse(commit.TreeHash())
		util.Check(err)
	}

	// "obj" is now a valid tree object.
	tree, err := git.NewTree(repo, obj)
	util.Check(err)

	fmt.Print(tree.Print())
}
