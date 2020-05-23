package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

// ShowRefCommand lists the components of "show-ref" comamnd.
type ShowRefCommand struct {
	fs       *flag.FlagSet
	showHead bool
	verify   bool
	pattern  string
}

// NewShowRefCommand creates a new command object.
func NewShowRefCommand() *ShowRefCommand {
	cmd := &ShowRefCommand{
		fs: flag.NewFlagSet("show-ref", flag.ExitOnError),
	}

	cmd.fs.BoolVar(&cmd.showHead, "head", false,
		"Show the HEAD reference, even if it would normally be filtered out.")
	cmd.fs.BoolVar(&cmd.verify, "verify", false,
		"Enable stricter reference checking by requiring an exact ref path.")

	return cmd
}

// Name gives the name of the command.
func (cmd *ShowRefCommand) Name() string {
	return cmd.fs.Name()
}

// Init initializes and validates the given command.
func (cmd *ShowRefCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() == 0 && cmd.verify {
		return errors.New("fatal: --verify requires a reference")
	}

	if cmd.fs.NArg() >= 1 {
		cmd.pattern = cmd.fs.Arg(0)
	}

	return nil
}

// Description gives the description of the command.
func (cmd *ShowRefCommand) Description() string {
	return "List references in a local repository"
}

// Usage prints the usage string for the end user.
func (cmd *ShowRefCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] <pattern/reference>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *ShowRefCommand) Execute() {
	repo, err := git.GetRepo(".")
	util.Check(err)

	if cmd.verify {
		refHash, err := repo.ValidateRef(cmd.pattern)
		util.Check(err)

		fmt.Printf("%s %s\n", refHash, cmd.pattern)
	} else {
		// Get a list of all the references.
		refs, err := repo.GetRefs(cmd.pattern, cmd.showHead)
		util.Check(err)

		for _, ref := range refs {
			fmt.Printf("%s %s\n", ref.RefHash, ref.Name)
		}
	}
}
