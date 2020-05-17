package main

import (
	"errors"
	"flag"
	"fmt"
	"gogit"
	"os"
)

type LogCommand struct {
	fs         *flag.FlagSet
	limit      uint
	commitHash string
}

func NewLogCommand() *LogCommand {
	cmd := &LogCommand{
		fs: flag.NewFlagSet("log", flag.ExitOnError),
	}

	cmd.fs.UintVar(&cmd.limit, "n", 0, "Limit the number of commits to output")
	return cmd
}

func (cmd *LogCommand) Name() string {
	return cmd.fs.Name()
}

func (cmd *LogCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		return errors.New("Error: Missing <commit-hash> argument\n")
	}

	cmd.commitHash = cmd.fs.Arg(0)
	return nil
}

func (cmd *LogCommand) Description() string {
	return "Shows the commit logs starting with given commit"
}

func (cmd *LogCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] <commit-hash>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

func (cmd *LogCommand) Execute() {
	repo, err := gogit.GetRepo(".")
	Check(err)

	// Resolve the given hash to a full hash.
	commitHash, err := repo.ObjectFind(cmd.commitHash)
	Check(err)

	var printed uint
	for {
		obj, err := repo.ObjectParse(commitHash)
		if err != nil || obj.ObjType != "commit" {
			fmt.Println("fatal: not a commit object.")
			os.Exit(1)
		}

		// Print this commit now.
		commit, err := gogit.NewCommit(obj)
		Check(err)
		fmt.Println(commit.Print())

		// See if the user specified limit is reached.
		printed += 1
		if cmd.limit > 0 && printed == cmd.limit {
			break
		}

		// Find the parent list of this commit.
		parents := commit.Parents()

		// If there are no more parents (base commit), then stop.
		if len(parents) == 0 {
			break
		}

		// Currently, "gogit log" only supports a single parent. In real "git",
		// there can be more than one parent in "merge" scenarios.
		commitHash = parents[0]
	}
}
