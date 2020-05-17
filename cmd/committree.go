package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/ssrathi/gogit/git"
)

type CommitTreeCommand struct {
	fs         *flag.FlagSet
	treeHash   string
	parentHash string
	msg        string
}

func NewCommitTreeCommand() *CommitTreeCommand {
	cmd := &CommitTreeCommand{
		fs: flag.NewFlagSet("commit-tree", flag.ExitOnError),
	}

	cmd.fs.StringVar(&cmd.parentHash, "p", "", "id of a parent commit object")
	cmd.fs.StringVar(&cmd.msg, "m", "", "A paragraph in the commit log message")
	return cmd
}

func (cmd *CommitTreeCommand) Name() string {
	return cmd.fs.Name()
}

func (cmd *CommitTreeCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		return errors.New("Error: Missing <tree> argument\n")
	}

	// Message is currently mandatory till getting it from an editor is implemented.
	if cmd.msg == "" {
		return errors.New("Error: Missing [-m message] argument\n")
	}

	cmd.treeHash = cmd.fs.Arg(0)
	return nil
}

func (cmd *CommitTreeCommand) Description() string {
	return "Create a new commit object"
}

func (cmd *CommitTreeCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] <tree>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

func (cmd *CommitTreeCommand) Execute() {
	repo, err := git.GetRepo(".")
	Check(err)

	// Add a new line to the msg.
	msg := cmd.msg + "\n"
	commit, err := git.NewCommitFromParams(
		cmd.treeHash, cmd.parentHash, msg)
	Check(err)

	// Write the commit now.
	hash, err := repo.ObjectWrite(commit.Obj, true)
	Check(err)

	fmt.Println(hash)
}
