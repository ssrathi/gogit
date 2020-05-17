package cmd

import (
	"flag"
	"fmt"

	"github.com/ssrathi/gogit/git"
)

// InitCommand lists the components of "init" comamnd.
type InitCommand struct {
	fs   *flag.FlagSet
	path string
}

// NewCommitTreeCommand creates a new command object.
func NewInitCommand() *InitCommand {
	cmd := &InitCommand{
		fs: flag.NewFlagSet("init", flag.ExitOnError),
	}

	cmd.fs.StringVar(&cmd.path, "path", ".", "Path to create the repository")
	return cmd
}

// Name gives the name of the command.
func (cmd *InitCommand) Name() string {
	return cmd.fs.Name()
}

// Description gives the description of the command.
func (cmd *InitCommand) Description() string {
	return "Create an empty Git repository"
}

// Init initializes and validates the given command.
func (cmd *InitCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	return cmd.fs.Parse(args)
}

// Usage prints the usage string for the end user.
func (cmd *InitCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>]\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *InitCommand) Execute() {
	repo, err := git.NewRepo(cmd.path)
	Check(err)

	fmt.Printf("Initialized empty Git repository in %s/\n", repo.GitDir)
}
