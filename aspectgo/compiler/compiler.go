// Package compiler provides the AspectGo compiler.
package compiler

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/exp/aspectgo/compiler/gopath"
	"golang.org/x/exp/aspectgo/compiler/parse"
	"golang.org/x/exp/aspectgo/compiler/weave"
)

// Compiler is the type for the AspectGo compiler.
type Compiler struct {
	// WovenGOPATH is the GOPATH for woven packages.
	WovenGOPATH string

	// Target is the target package name.
	// Can contain ... for recursive weaving.
	Target string

	// AspectFilenames are aspect file names.
	// currently, only single aspect file is supported
	AspectFilenames []string
}

// Do does all the compilation phases.
func (c *Compiler) Do() error {
	log.Printf("Phase 0: Checking arguments")
	if c.WovenGOPATH == "" {
		return errors.New("WovenGOPATH not specified")
	}
	if c.Target == "" {
		return errors.New("Target not specified")
	}
	if len(c.AspectFilenames) != 1 {
		return fmt.Errorf("only single aspect file is supported at the moment: %v", c.AspectFilenames)
	}
	aspectFilename := c.AspectFilenames[0]
	oldGOPATH := os.Getenv("GOPATH")
	if oldGOPATH == "" {
		return errors.New("GOPATH not set")
	}

	log.Printf("Phase 1: Parsing the aspects")
	aspectFile, err := parse.ParseAspectFile(aspectFilename)
	if err != nil {
		return err
	}

	log.Printf("Phase 2: Weaving the aspects to the target packages")
	targets, err := resolveTarget(oldGOPATH, c.Target)
	if err != nil {
		return err
	}
	var writtenFnames []string
	for _, target := range targets {
		w, err := weave.Weave(c.WovenGOPATH, target, aspectFile)
		if err != nil {
			return err
		}
		writtenFnames = append(writtenFnames, w...)
	}
	if len(writtenFnames) == 0 {
		log.Printf("Nothing to do")
		return nil
	}

	log.Printf("Phase 3: Fixing up GOPATH")
	err = gopath.FixUp(oldGOPATH, c.WovenGOPATH, writtenFnames)
	if err != nil {
		return err
	}
	return nil
}

// resolveTarget resolves target that can contain ... and returns the list of
// resolved packages.
func resolveTarget(gopath, target string) ([]string, error) {
	if strings.HasPrefix(target, ".") {
		return nil, errors.New("local package (.) is not supported yet")
	}
	if filepath.Base(target) != "..." {
		return []string{target}, nil
	}
	gopathSrc := filepath.Join(gopath, "src")
	d := filepath.Join(gopathSrc, filepath.Dir(target))
	resolvedMap := make(map[string]struct{})
	err := filepath.Walk(d,
		func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, ".go") {
				d := filepath.Dir(path)
				_, ok := resolvedMap[d]
				if !ok {
					resolvedMap[d] = struct{}{}
				}
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	var resolved []string
	for f, _ := range resolvedMap {
		resolved = append(resolved, strings.Replace(f, gopathSrc+"/", "", 1))
	}
	sort.Strings(resolved)
	return resolved, nil
}
