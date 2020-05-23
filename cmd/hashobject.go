package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

// HashObjectCommand lists the components of "cat-file" comamnd.
type HashObjectCommand struct {
	fs    *flag.FlagSet
	write bool
	file  string
}

// NewHashObjectCommand creates a new command object.
func NewHashObjectCommand() *HashObjectCommand {
	cmd := &HashObjectCommand{
		fs: flag.NewFlagSet("hash-object", flag.ExitOnError),
	}

	cmd.fs.BoolVar(&cmd.write, "w", false,
		"Actually write the object into the object database.")

	return cmd
}

// Name gives the name of the command.
func (cmd *HashObjectCommand) Name() string {
	return cmd.fs.Name()
}

// Description gives the description of the command.
func (cmd *HashObjectCommand) Description() string {
	return "Compute object ID and optionally creates a blob from a file"
}

// Init initializes and validates the given command.
func (cmd *HashObjectCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		return errors.New("Error: Missing <file> argument")
	}

	cmd.file = cmd.fs.Arg(0)
	return nil
}

// Usage prints the usage string for the end user.
func (cmd *HashObjectCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] <file>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *HashObjectCommand) Execute() {
	repo, err := git.GetRepo(".")
	util.Check(err)

	blob, err := git.NewBlobFromFile(repo, cmd.file)
	util.Check(err)

	sha1, err := repo.ObjectWrite(blob.Obj, cmd.write)
	util.Check(err)

	fmt.Println(sha1)
}
