package main

import (
	"os"

	"golang.org/x/exp/aspectgo/compiler/cli"
)

func main() {
	os.Exit(cli.Main(os.Args))
}
