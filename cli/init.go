package main

import (
	"flag"
	"fmt"
	"gogit"
)

type InitCommand struct {
	fs   *flag.FlagSet
	path string
}

func NewInitCommand() *InitCommand {
	cmd := &InitCommand{
		fs: flag.NewFlagSet("init", flag.ExitOnError),
	}

	cmd.fs.StringVar(&cmd.path, "path", ".", "Path to create the repository")
	return cmd
}

func (cmd *InitCommand) Name() string {
	return cmd.fs.Name()
}

func (cmd *InitCommand) Description() string {
	return "Create an empty Git repository"
}

func (cmd *InitCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	return cmd.fs.Parse(args)
}

func (cmd *InitCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>]\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

func (cmd *InitCommand) Validate() error {
	// There is nothing to validate for 'init' command.
	return nil
}

func (cmd *InitCommand) Execute() {
	repo, err := gogit.NewRepo(cmd.path)
	Check(err)

	fmt.Printf("Initialized empty Git repository in %s/\n", repo.GitDir)
}
