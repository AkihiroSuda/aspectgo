// Package util provides miscellaneous utilities.
package util

import (
	"bufio"
	"bytes"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
)

// DebugMode denotes the debug flag.
var DebugMode = false

// LocalTypeString returns the local type string for typ, using pkg and imports information.
// TODO: use regexp.
func LocalTypeString(typ types.Type, pkg *types.Package, imports []*ast.ImportSpec) (string, error) {
	full := types.TypeString(typ,
		types.RelativeTo(pkg))

	for _, imp := range imports {
		// imp.Path.Value contains double-quotes. so we move it first.
		pathValue := strings.Replace(imp.Path.Value, "\"", "", -1)

		matched := strings.Contains(full, pathValue)
		if !matched {
			continue
		}

		replacement := filepath.Base(pathValue)
		if imp.Name != nil {
			if imp.Name.Name == "." {
				replacement = ""
			} else {
				replacement = imp.Name.Name
			}
		}
		replaced := strings.Replace(full, pathValue, replacement, -1)
		return replaced, nil
	}
	// no import spec matched. this is expected for built-in types.
	return full, nil
}

// ASTDebugString returns a debug string for the AST node.
func ASTDebugString(node ast.Node) string {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := ast.Fprint(w, token.NewFileSet(), node, nil)
	if err != nil {
		return err.Error()
	}
	w.Flush()
	s := string(b.Bytes())
	return "\n" + s
}
