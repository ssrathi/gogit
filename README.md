[![Go Report Card](https://goreportcard.com/badge/github.com/ssrathi/gogit)](https://goreportcard.com/report/github.com/ssrathi/gogit)
[![GoDoc](https://godoc.org/github.com/ssrathi/gogit?status.svg)](https://godoc.org/github.com/ssrathi/gogit)
# gogit

Implementation of git internal commands in Go language.

This project is part of a learning exercise to implement a subset of "git"
commands. It can be used to create and maintain git objects, such as blobs,
trees, commits, branches and tags.

[![asciicast](https://asciinema.org/a/331278.svg)](https://asciinema.org/a/331278?speed=2)

## Supported commands
```
gogit - the stupid content tracker

usage: gogit <command> [<args>]
Valid commands:
  init           Create an empty Git repository
  hash-object    Compute object ID and optionally creates a blob from a file
  cat-file       Provide content or type and size information for repository objects
  ls-tree        List the contents of a tree object
  mktree         Build a tree-object from ls-tree formatted text
  checkout       restore working tree files
  commit-tree    Create a new commit object
  log            Shows the commit logs

Use "gogit <command> --help" for help on a specific command
```
