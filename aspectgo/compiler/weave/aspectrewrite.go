package weave

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/format"
	"log"
	"os"
	"path/filepath"

	rewrite "github.com/tsuna/gorewrite"

	"golang.org/x/tools/go/loader"

	"golang.org/x/exp/aspectgo/compiler/consts"
	"golang.org/x/exp/aspectgo/compiler/parse"
)

func rewriteAspectFile(wovenGOPATH string, af *parse.AspectFile) ([]string, error) {
	// look up *ast.File object
	var target *ast.File
	for _, file := range af.PkgInfo.Files {
		posn := af.Program.Fset.Position(file.Pos())
		if posn.Filename == af.Filename {
			target = file
		}
	}
	if target == nil {
		return nil, fmt.Errorf("No ast node found for %s", af.Filename)
	}
	// prepare file name
	wovenPkgPath := filepath.Join(filepath.Join(wovenGOPATH, "src"),
		"agaspect")
	err := os.MkdirAll(wovenPkgPath, 0755)
	if err != nil {
		return nil, err
	}
	outFilename := filepath.Join(wovenPkgPath, "aspect.go")
	outFile, err := os.Create(outFilename)
	if err != nil {
		return nil, err
	}
	defer outFile.Close()

	// rewrite
	log.Printf("Rewriting aspect file %s --> %s", af.Filename, outFilename)
	rw := &aspectFileRewriter{
		Program: af.Program,
	}
	rewritten := rewrite.Rewrite(rw, target)

	// write the buffer
	outW := bufio.NewWriter(outFile)
	outW.Write([]byte(consts.AutogenFileHeader))
	format.Node(outW, af.Program.Fset, rewritten)
	outW.Flush()
	return []string{outFilename}, nil
}

// aspectFileRewriter implements rewrite.Rewriter
type aspectFileRewriter struct {
	Program *loader.Program
}

func (r *aspectFileRewriter) Rewrite(node ast.Node) (ast.Node, rewrite.Rewriter) {
	switch n := node.(type) {
	case *ast.File:
		oldName := n.Name.Name
		if oldName != "main" {
			log.Fatalf("impl error: why not main? this is unexpected and critical: %s: %v", oldName, n)
		}
		newName := "agaspect"
		rewritten := *n
		rewritten.Name = ast.NewIdent(newName)
		return &rewritten, r
	}
	return node, r
}
