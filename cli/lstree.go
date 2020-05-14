package main

import (
	"errors"
	"flag"
	"fmt"
	"gogit"
	"os"
)

type LsTreeCommand struct {
	fs      *flag.FlagSet
	objHash string
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
	return cmd.fs.Parse(args)
}

func (cmd *LsTreeCommand) Description() string {
	return "List the contents of a tree object"
}

func (cmd *LsTreeCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s <tree-ish>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

func (cmd *LsTreeCommand) Validate() error {
	if cmd.fs.NArg() < 1 {
		return errors.New("Error: Missing <tree-ish> argument\n")
	}

	cmd.objHash = cmd.fs.Arg(0)
	return nil
}

func (cmd *LsTreeCommand) Execute() {
	repo, err := gogit.GetRepo(".")
	Check(err)

	obj, err := repo.ObjectParse(cmd.objHash)
	if err != nil || obj.ObjType != "tree" {
		fmt.Println("fatal: not a tree object")
		os.Exit(1)
	}

	tree, err := gogit.NewTree(obj)
	Check(err)

	fmt.Print(tree.Print())
}
