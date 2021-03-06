package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

// LogCommand lists the components of "log" comamnd.
type LogCommand struct {
	fs       *flag.FlagSet
	limit    uint
	revision string
}

// NewLogCommand creates a new command object.
func NewLogCommand() *LogCommand {
	cmd := &LogCommand{
		fs: flag.NewFlagSet("log", flag.ExitOnError),
	}

	cmd.fs.UintVar(&cmd.limit, "n", 0, "Limit the number of commits to output")
	return cmd
}

// Name gives the name of the command.
func (cmd *LogCommand) Name() string {
	return cmd.fs.Name()
}

// Init initializes and validates the given command.
func (cmd *LogCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		cmd.revision = "HEAD"
	} else {
		cmd.revision = cmd.fs.Arg(0)
	}

	return nil
}

// Description gives the description of the command.
func (cmd *LogCommand) Description() string {
	return "Shows the commit logs"
}

// Usage prints the usage string for the end user.
func (cmd *LogCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] [<revision>]\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *LogCommand) Execute() {
	repo, err := git.GetRepo(".")
	util.Check(err)

	// Resolve the given revision to a full hash.
	commitHash, err := repo.UniqueNameResolve(cmd.revision)
	util.Check(err)

	var printed uint
	for {
		obj, err := repo.ObjectParse(commitHash)
		if err != nil || obj.ObjType != "commit" {
			fmt.Printf("fatal: not a commit object (%s)\n", commitHash)
			os.Exit(1)
		}

		// Print this commit now.
		commit, err := git.NewCommit(repo, obj)
		util.Check(err)
		commitStr, err := commit.PrettyPrint()
		util.Check(err)

		// Print the commit msg now. If it doesn't end with a newline, then
		// add one manually.
		fmt.Printf(commitStr)
		if commitStr[len(commitStr)-1] != byte('\n') {
			fmt.Println()
		}

		// See if the user specified limit is reached.
		printed++
		if cmd.limit > 0 && printed == cmd.limit {
			break
		}

		// Find the parent list of this commit.
		parents := commit.Parents()
		// If there are no more parents (base commit), then stop.
		if len(parents) == 0 {
			break
		}

		// Currently, "git log" only supports a single parent. In real "git",
		// there can be more than one parent in "merge" scenarios.
		commitHash = parents[0]

		// Put a new line between two successive commits.
		fmt.Println()
	}
}
