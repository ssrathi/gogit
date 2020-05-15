package main

import (
	"errors"
	"flag"
	"fmt"
	"gogit"
)

type HashObjectCommand struct {
	fs    *flag.FlagSet
	write bool
	file  string
}

func NewHashObjectCommand() *HashObjectCommand {
	cmd := &HashObjectCommand{
		fs: flag.NewFlagSet("hash-object", flag.ExitOnError),
	}

	cmd.fs.BoolVar(&cmd.write, "w", false,
		"Actually write the object into the object database.")

	return cmd
}

func (cmd *HashObjectCommand) Name() string {
	return cmd.fs.Name()
}

func (cmd *HashObjectCommand) Description() string {
	return "Compute object ID and optionally creates a blob from a file"
}

func (cmd *HashObjectCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		return errors.New("Error: Missing <file> argument\n")
	}

	cmd.file = cmd.fs.Arg(0)
	return nil
}

func (cmd *HashObjectCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] <file>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

func (cmd *HashObjectCommand) Execute() {
	repo, err := gogit.GetRepo(".")
	Check(err)

	blob, err := gogit.NewBlobFromFile(cmd.file)
	Check(err)

	sha1, err := repo.ObjectWrite(blob.Obj, cmd.write)
	Check(err)

	fmt.Println(sha1)
}
