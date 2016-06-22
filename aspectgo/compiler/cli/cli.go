// Package cli provides the CLI for AspectGo.
package cli

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/exp/aspectgo/compiler"
	"golang.org/x/exp/aspectgo/compiler/util"
)

// Main is the CLI for AspectGo.
func Main(args []string) int {
	var (
		debug  bool
		weave  string
		target string
	)
	f := flag.NewFlagSet(args[0], flag.ExitOnError)
	f.BoolVar(&debug, "debug", false, "enable debug print")
	f.StringVar(&weave, "w", "/tmp/wovengopath", "woven gopath")
	f.StringVar(&target, "t", "", "target package name")
	f.Parse(args[1:])

	if target == "" {
		fmt.Fprintf(os.Stderr, "No target package specified\n")
		return 1
	}
	if f.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "No aspect file specified\n")
		return 1
	}
	if f.NArg() >= 2 {
		fmt.Fprintf(os.Stderr, "Too many aspect files specified: %s\n", f.Args())
		return 1
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	util.DebugMode = debug
	if util.DebugMode {
		log.Printf("running in debug mode")
	}

	aspectFile := f.Args()[0]
	comp := compiler.Compiler{
		WovenGOPATH:     weave,
		Target:          target,
		AspectFilenames: []string{aspectFile},
	}
	if err := comp.Do(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
