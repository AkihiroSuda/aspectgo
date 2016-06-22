package main

import (
	"os"

	"github.com/AkihiroSuda/aspectgo/compiler/cli"
)

func main() {
	os.Exit(cli.Main(os.Args))
}
