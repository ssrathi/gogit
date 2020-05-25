package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

// CatFileCommand lists the components of "cat-file" comamnd.
type CatFileCommand struct {
	fs       *flag.FlagSet
	revision string
	getType  bool
	getSize  bool
	printObj bool
}

// NewCatFileCommand creates a new command object.
func NewCatFileCommand() *CatFileCommand {
	cmd := &CatFileCommand{
		fs: flag.NewFlagSet("cat-file", flag.ExitOnError),
	}

	cmd.fs.BoolVar(&cmd.getType, "t", false,
		"Instead of the content, show the object type identified by <object>")
	cmd.fs.BoolVar(&cmd.getSize, "s", false,
		"Instead of the content, show the object size identified by <object>")
	cmd.fs.BoolVar(&cmd.printObj, "p", false,
		"Pretty-print the contents of <object> based on its type.")
	return cmd
}

// Name gives the name of the command.
func (cmd *CatFileCommand) Name() string {
	return cmd.fs.Name()
}

// Description gives the description of the command.
func (cmd *CatFileCommand) Description() string {
	return "Provide content or type and size information for repository objects"
}

// Init initializes and validates the given command.
func (cmd *CatFileCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		return errors.New("error: Missing <object> argument")
	}

	// All optional boolean args are mutually exclusive
	if !(cmd.getType || cmd.getSize || cmd.printObj) {
		return errors.New("error: one of '-t', '-s' or '-p' must be provided")
	}

	if cmd.getType && cmd.getSize {
		return errors.New("error: switch 't' and 's' are incompatible")
	}
	if cmd.getSize && cmd.printObj {
		return errors.New("error: switch 's' and 'p' are incompatible")
	}
	if cmd.printObj && cmd.getType {
		return errors.New("error: switch 'p' and 't' are incompatible")
	}

	cmd.revision = cmd.fs.Arg(0)
	return nil
}

// Usage prints the usage string for the end user.
func (cmd *CatFileCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] <object>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *CatFileCommand) Execute() {
	repo, err := git.GetRepo(".")
	util.Check(err)

	// Resolve the given name to a full hash.
	objHash, err := repo.UniqueNameResolve(cmd.revision)
	util.Check(err)

	obj, err := repo.ObjectParse(objHash)
	util.Check(err)

	var objIntf git.ObjIntf
	switch obj.ObjType {
	case "blob":
		objIntf, err = git.NewBlob(repo, obj)
		util.Check(err)
	case "tree":
		objIntf, err = git.NewTree(repo, obj)
		util.Check(err)
	case "commit":
		objIntf, err = git.NewCommit(repo, obj)
		util.Check(err)
	}

	// Only one of 'printObj', 'getType' and 'getSize' is provided.
	if cmd.printObj {
		fmt.Print(objIntf.Print())
	} else if cmd.getType {
		fmt.Println(objIntf.Type())
	} else if cmd.getSize {
		fmt.Println(objIntf.DataSize())
	}
}
