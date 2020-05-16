[![Go Report Card](https://goreportcard.com/badge/github.com/ssrathi/golang_git)](https://goreportcard.com/report/github.com/ssrathi/golang_git)
[![GoDoc](https://godoc.org/github.com/ssrathi/golang_git?status.svg)](https://godoc.org/github.com/ssrathi/golang_git)
# golang_git

Implementation of git internal commands in Go language.

## Supported commands
```
gogit - the stupid content tracker

usage: ./gogit <command> [<args>]
Valid commands:
  init           Create an empty Git repository
  hash-object    Compute object ID and optionally creates a blob from a file
  cat-file       Provide content or type and size information for repository objects
  ls-tree        List the contents of a tree object
  mktree         Build a tree-object from ls-tree formatted text
  checkout       restore working tree files
  commit-tree    Create a new commit object
  log            Shows the commit logs starting with given commit
```
