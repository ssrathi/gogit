package main

import (
	"flag"
	"fmt"
	"os"
)

func cmd_init(repo_path string) {
	fmt.Printf("Init done at path: %q\n", repo_path)
}

func cmd_hash_object() {
	fmt.Println("hash-object done")
}

func cmd_cat_file() {
	fmt.Println("cat-file done")
}

func Usage() {
	fmt.Printf("usage: %s <command> [<args>]\n", os.Args[0])
	fmt.Println("Valid commands:")
	fmt.Println("  init         Inititialize an empty Git repository")
	fmt.Println("  hash-object  Compute object ID and optionally create a blob")
	fmt.Println("  cat-file     Get content information for a Git object")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage

	// Subcommands
	initCommand := flag.NewFlagSet("init", flag.ExitOnError)
	hashObjectCommand := flag.NewFlagSet("hash-object", flag.ExitOnError)
	catFileCommand := flag.NewFlagSet("cat-file", flag.ExitOnError)

	// Options for 'init' subcommand
	initPath := initCommand.String("path", ".", "Path to create the repository")

	if len(os.Args) < 2 {
		Usage()
		os.Exit(1)
	}

	flag.Parse()
	switch os.Args[1] {
	case "init":
		initCommand.Parse(os.Args[2:])
		cmd_init(*initPath)
	case "hash-object":
		hashObjectCommand.Parse(os.Args[2:])
		cmd_hash_object()
	case "cat-file":
		catFileCommand.Parse(os.Args[2:])
		cmd_cat_file()
	default:
		Usage()
		os.Exit(1)
	}
}
