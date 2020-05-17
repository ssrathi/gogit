package main

import (
	"os"

	"github.com/ssrathi/gogit/cmd"
)

func main() {
	cmd.Execute(os.Args[0], os.Args[1:])
}
