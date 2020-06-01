[![Builds](https://github.com/ssrathi/gogit/workflows/Build/badge.svg?branch=master)](https://github.com/ssrathi/gogit/actions?query=branch%3Amaster+workflow%3ABuild)
[![Go Report Card](https://goreportcard.com/badge/github.com/ssrathi/gogit)](https://goreportcard.com/report/github.com/ssrathi/gogit)
[![GoDoc](https://godoc.org/github.com/ssrathi/gogit?status.svg)](https://godoc.org/github.com/ssrathi/gogit)

<img src="https://raw.githubusercontent.com/ssrathi/gogit/master/assets/cover.png" width="200"/>

# gogit

Implementation of git internal commands in Go language.

This project is part of a learning exercise to implement a subset of "git"
commands. It can be used to create and maintain git objects, such as blobs,
trees, commits, references and tags.

[<img src="https://asciinema.org/a/331278.svg" alt="demo" width="400" height="240"/>](https://asciinema.org/a/331278?speed=2&autoplay=1&t=8)

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
  show-ref       List references in a local repository
  update-ref     Update the object name stored in a ref safely
  rev-parse      Parse a given git identifier

Use "gogit <command> --help" for help on a specific command
```

## Installation
```
go get github.com/ssrathi/gogit
```

## Contributing

Contributions are most welcome! Please follow the steps below to send
pull requests with your changes.

* Fork this repository and create a feature branch in it.
* Push a commit with your changes.
* Create a new pull request.
* Create a new issue and link the pull request to it.

