package main

import (
	"errors"
	"flag"
	"fmt"
	"gogit"
	"os"
)

type CheckoutCommand struct {
	fs      *flag.FlagSet
	path    string
	objHash string
}

func NewCheckoutCommand() *CheckoutCommand {
	cmd := &CheckoutCommand{
		fs: flag.NewFlagSet("checkout", flag.ExitOnError),
	}

	cmd.fs.StringVar(&cmd.path, "path", ".", "Path to create the files")
	return cmd
}

func (cmd *CheckoutCommand) Name() string {
	return cmd.fs.Name()
}

func (cmd *CheckoutCommand) Description() string {
	return "restore working tree files"
}

func (cmd *CheckoutCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		return errors.New("Error: Missing <object> argument\n")
	}

	cmd.objHash = cmd.fs.Arg(0)
	return nil
}

func (cmd *CheckoutCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] <object>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

func (cmd *CheckoutCommand) Execute() {
	repo, err := gogit.GetRepo(".")
	Check(err)

	obj, err := repo.ObjectParse(cmd.objHash)
	if err != nil || obj.ObjType != "tree" {
		fmt.Println("fatal: not a tree object")
		os.Exit(1)
	}

	tree, err := gogit.NewTree(obj)
	Check(err)

	err = tree.Checkout(cmd.path)
	Check(err)
}
