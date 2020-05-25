package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

// CommitTreeCommand lists the components of "commit-tree" comamnd.
type CommitTreeCommand struct {
	fs         *flag.FlagSet
	treeHash   string
	parentHash string
	msg        string
}

// NewCommitTreeCommand creates a new command object.
func NewCommitTreeCommand() *CommitTreeCommand {
	cmd := &CommitTreeCommand{
		fs: flag.NewFlagSet("commit-tree", flag.ExitOnError),
	}

	cmd.fs.StringVar(&cmd.parentHash, "p", "", "id of a parent commit object")
	cmd.fs.StringVar(&cmd.msg, "m", "", "A paragraph in the commit log message")
	return cmd
}

// Name gives the name of the command.
func (cmd *CommitTreeCommand) Name() string {
	return cmd.fs.Name()
}

// Init initializes and validates the given command.
func (cmd *CommitTreeCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		return errors.New("error: Missing <tree> argument")
	}

	// Message is currently mandatory till getting it from an editor is implemented.
	if cmd.msg == "" {
		return errors.New("error: Missing [-m message] argument")
	}

	cmd.treeHash = cmd.fs.Arg(0)
	return nil
}

// Description gives the description of the command.
func (cmd *CommitTreeCommand) Description() string {
	return "Create a new commit object"
}

// Usage prints the usage string for the end user.
func (cmd *CommitTreeCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] <tree>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *CommitTreeCommand) Execute() {
	repo, err := git.GetRepo(".")
	util.Check(err)

	// Add a new line to the msg.
	msg := cmd.msg + "\n"
	commit, err := git.NewCommitFromParams(
		repo, cmd.treeHash, cmd.parentHash, msg)
	util.Check(err)

	// Write the commit now.
	hash, err := repo.ObjectWrite(commit.Object, true)
	util.Check(err)

	fmt.Println(hash)
}
