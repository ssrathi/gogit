// CLI entry point for gogit binary.
package main

import (
	"flag"
	"fmt"
	"gogit"
	"io/ioutil"
	"log"
	"os"
)

func cmd_init(repo_path string) {
	repo, err := gogit.NewRepo(repo_path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Initialized empty Git repository in %s/\n", repo.GitDir)
}

func cmd_hash_object(file string, write bool) {
	repo, err := gogit.GetRepo(".")
	if err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	blob := gogit.NewBlob(repo, data)
	sha1, err := blob.Write(write)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(sha1)
}

func cmd_cat_file(objHash string, getType bool, getSize bool, printObj bool) {
	repo, err := gogit.GetRepo(".")
	if err != nil {
		log.Fatal(err)
	}

	blob := gogit.NewBlob(repo, nil)
	if err := blob.Parse(objHash); err != nil {
		log.Fatal(err)
	}

	// Only one of 'printObj', 'getType' and 'getSize' is provided.
	if printObj {
		fmt.Print(string(blob.Data))
	} else if getType {
		fmt.Println(blob.ObjType)
	} else if getSize {
		fmt.Println(len(blob.Data))
	}
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
	if len(os.Args) < 2 {
		Usage()
		os.Exit(1)
	}

	// Subcommands
	initCommand := flag.NewFlagSet("init", flag.ExitOnError)
	hashObjectCommand := flag.NewFlagSet("hash-object", flag.ExitOnError)
	catFileCommand := flag.NewFlagSet("cat-file", flag.ExitOnError)

	// Options for 'init' subcommand
	initPath := initCommand.String("path", ".", "Path to create the repository")

	// Options for 'hash-object' subcommand
	hashObjectWriteObj := hashObjectCommand.Bool("w", false,
		"Actually write the object into the object database.")

	// Options for 'cat-file' subcommand
	catFileGetType := catFileCommand.Bool("t", false,
		"Instead of the content, show the object type identified by <object>")
	catFileGetSize := catFileCommand.Bool("s", false,
		"Instead of the content, show the object size identified by <object>")
	catFilePrint := catFileCommand.Bool("p", false,
		"Pretty-print the contents of <object> based on its type.")

	flag.Parse()
	switch os.Args[1] {
	case "init":
		initCommand.Parse(os.Args[2:])

		// Execute the command.
		cmd_init(*initPath)

	case "hash-object":
		hashObjectCommand.Parse(os.Args[2:])
		if hashObjectCommand.NArg() != 1 {
			hashObjectCommand.Usage()
			os.Exit(1)
		}

		// Execute the command.
		cmd_hash_object(hashObjectCommand.Arg(0), *hashObjectWriteObj)

	case "cat-file":
		catFileCommand.Parse(os.Args[2:])
		if catFileCommand.NArg() != 1 {
			catFileCommand.Usage()
			os.Exit(1)
		}

		// Execute the command.
		cmd_cat_file(catFileCommand.Arg(0), *catFileGetType, *catFileGetSize,
			*catFilePrint)

	default:
		Usage()
		os.Exit(1)
	}
}
