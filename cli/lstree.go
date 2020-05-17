package main

import (
	"errors"
	"flag"
	"fmt"
	"gogit"
	"os"
)

type LsTreeCommand struct {
	fs       *flag.FlagSet
	revision string
}

func NewLsTreeCommand() *LsTreeCommand {
	cmd := &LsTreeCommand{
		fs: flag.NewFlagSet("ls-tree", flag.ExitOnError),
	}
	return cmd
}

func (cmd *LsTreeCommand) Name() string {
	return cmd.fs.Name()
}

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

func (cmd *LsTreeCommand) Description() string {
	return "List the contents of a tree object"
}

func (cmd *LsTreeCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s <tree-ish>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

func (cmd *LsTreeCommand) Execute() {
	repo, err := gogit.GetRepo(".")
	Check(err)

	// Resolve the given hash to a full hash.
	objHash, err := repo.ObjectFind(cmd.revision)
	Check(err)

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
		commit, err := gogit.NewCommit(obj)
		Check(err)
		obj, err = repo.ObjectParse(commit.TreeHash())
		Check(err)
	}

	// "obj" is now a valid tree object.
	tree, err := gogit.NewTree(obj)
	Check(err)

	fmt.Print(tree.Print())
}
