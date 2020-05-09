// CLI entry point for gogit binary.
package main

import (
	"flag"
	"fmt"
	"gogit"
	"os"
)

func cmd_init(repo_path string) {
	repo, err := gogit.NewRepo(repo_path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Initialized empty Git repository in %s/\n", repo.GitDir)
}

func cmd_hash_object() {
	repo, err := gogit.GetRepo(".")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Repo found at %q\n", repo.GitDir)
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
