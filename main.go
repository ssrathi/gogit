/*
Implementation of git internal commands in Go language.

This project is part of a learning exercise to implement a subset of "git"
commands. It can be used to create and maintain git objects, such as blobs,
trees, commits, branches and tags.

Code Organization

 - "git": internal git objects and related APIs.
 - "cmd": Command line parsing and execution.
 - "util": Miscellaneous utility APIs.
*/
package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/ssrathi/gogit/cmd"
)

func init() {
	// Enable logging only if a specific ENV variable is set.
	if os.Getenv("GOGIT_DBG") != "1" {
		log.SetOutput(ioutil.Discard)
		log.SetFlags(0)
	} else {
		// Print file and line numbers in each log line.
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}
}

func main() {
	cmd.Execute()
}
