// CLI entry point for gogit binary.
package main

import (
	"flag"
	"fmt"
	"gogit"
	"io/ioutil"
	"os"
)

func cmdInit(path string) {
	repo, err := gogit.NewRepo(path)
	gogit.DieOnError(err)

	fmt.Printf("Initialized empty Git repository in %s/\n", repo.GitDir)
}

func cmdHashObject(file string, write bool) {
	repo, err := gogit.GetRepo(".")
	gogit.DieOnError(err)

	blob, err := gogit.NewBlobFromFile(file)
	gogit.DieOnError(err)

	sha1, err := repo.ObjectWrite(blob.Obj, write)
	gogit.DieOnError(err)

	fmt.Println(sha1)
}

func cmdCatFile(objHash string, getType bool, getSize bool, printObj bool) {
	repo, err := gogit.GetRepo(".")
	gogit.DieOnError(err)

	obj, err := repo.ObjectParse(objHash)
	gogit.DieOnError(err)

	var gitType gogit.GitType
	switch obj.ObjType {
	case "blob":
		gitType, err = gogit.NewBlob(obj)
		gogit.DieOnError(err)
	case "tree":
		gitType, err = gogit.NewTree(obj)
		gogit.DieOnError(err)
	}

	// Only one of 'printObj', 'getType' and 'getSize' is provided.
	if printObj {
		fmt.Print(gitType.Print())
	} else if getType {
		fmt.Println(gitType.Type())
	} else if getSize {
		fmt.Println(gitType.DataSize())
	}
}

func cmdLsTree(objHash string) {
	repo, err := gogit.GetRepo(".")
	gogit.DieOnError(err)

	obj, err := repo.ObjectParse(objHash)
	if err != nil || obj.ObjType != "tree" {
		fmt.Println("fatal: not a tree object")
		os.Exit(1)
	}

	tree, err := gogit.NewTree(obj)
	gogit.DieOnError(err)

	fmt.Print(tree.Print())
}

func cmdMkTree() {
	repo, err := gogit.GetRepo(".")
	gogit.DieOnError(err)

	input, err := ioutil.ReadAll(os.Stdin)
	gogit.DieOnError(err)

	tree, err := gogit.NewTreeFromInput(string(input))
	gogit.DieOnError(err)

	// Write the tree now.
	hash, err := repo.ObjectWrite(tree.Obj, true)
	gogit.DieOnError(err)

	fmt.Println(hash)
}

func cmdCheckout(objHash string, path string) {
	repo, err := gogit.GetRepo(".")
	gogit.DieOnError(err)

	obj, err := repo.ObjectParse(objHash)
	if err != nil || obj.ObjType != "tree" {
		fmt.Println("fatal: not a tree object")
		os.Exit(1)
	}

	tree, err := gogit.NewTree(obj)
	gogit.DieOnError(err)

	err = tree.Checkout(path)
	gogit.DieOnError(err)
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
	lsTreeCommand := flag.NewFlagSet("ls-tree", flag.ExitOnError)
	mkTreeCommand := flag.NewFlagSet("mktree", flag.ExitOnError)
	checkoutCommand := flag.NewFlagSet("checkout", flag.ExitOnError)

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

	// Options for 'checkout' subcommand
	checkoutPath := checkoutCommand.String("path", ".", "Path to checkout")

	flag.Parse()
	switch os.Args[1] {
	case "init":
		initCommand.Parse(os.Args[2:])

		// Execute the command.
		cmdInit(*initPath)

	case "hash-object":
		hashObjectCommand.Parse(os.Args[2:])
		if hashObjectCommand.NArg() != 1 {
			hashObjectCommand.Usage()
			os.Exit(1)
		}

		// Execute the command.
		cmdHashObject(hashObjectCommand.Arg(0), *hashObjectWriteObj)

	case "cat-file":
		catFileCommand.Parse(os.Args[2:])
		if catFileCommand.NArg() != 1 {
			catFileCommand.Usage()
			os.Exit(1)
		}

		// Execute the command.
		cmdCatFile(catFileCommand.Arg(0), *catFileGetType, *catFileGetSize,
			*catFilePrint)

	case "ls-tree":
		lsTreeCommand.Parse(os.Args[2:])
		if lsTreeCommand.NArg() != 1 {
			lsTreeCommand.Usage()
			os.Exit(1)
		}

		// Execute the command.
		cmdLsTree(lsTreeCommand.Arg(0))

	case "mktree":
		mkTreeCommand.Parse(os.Args[2:])

		// Execute the command.
		cmdMkTree()

	case "checkout":
		checkoutCommand.Parse(os.Args[2:])
		if checkoutCommand.NArg() != 1 {
			checkoutCommand.Usage()
			os.Exit(1)
		}

		// Execute the command.
		cmdCheckout(checkoutCommand.Arg(0), *checkoutPath)

	default:
		Usage()
		os.Exit(1)
	}
}
