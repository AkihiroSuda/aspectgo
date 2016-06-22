// Package parse provides the aspect parser.
package parse

import (
	"fmt"
	"go/parser"
	"go/types"

	"golang.org/x/tools/go/loader"

	"golang.org/x/exp/aspectgo/aspect"
	"golang.org/x/exp/aspectgo/compiler/consts"
)

const aspectPackagePath = consts.AspectGoPackagePath + "/aspect"

// AspectFile is the type for an aspect file.
type AspectFile struct {
	Filename  string
	Program   *loader.Program
	PkgInfo   *loader.PackageInfo
	Pointcuts map[*types.Named]aspect.Pointcut
}

// ParseAspectFile parses an aspect file.
func ParseAspectFile(aspectFilename string) (*AspectFile, error) {
	prog, pkgInfo, err := _parseAspectFile(aspectFilename)
	if err != nil {
		return nil, err
	}
	pkg := pkgInfo.Pkg
	if pkg.Name() != "main" {
		return nil, fmt.Errorf("aspect package name must be main: %s", pkg.Name())
	}
	aspectIntf, err := lookupAspectInterface(prog)
	if err != nil {
		return nil, err
	}
	aspects, err := lookupAspects(pkg, aspectIntf)
	if err != nil {
		return nil, err
	}
	aspectFile := &AspectFile{
		Filename:  aspectFilename,
		Program:   prog,
		PkgInfo:   pkgInfo,
		Pointcuts: make(map[*types.Named]aspect.Pointcut),
	}
	err = aspectFile.determinePointcuts(aspects)
	if err != nil {
		return nil, err
	}
	return aspectFile, nil
}

func _parseAspectFile(aspectFilename string) (*loader.Program, *loader.PackageInfo, error) {
	conf := loader.Config{
		ParserMode: parser.ParseComments,
	}
	conf.CreateFromFilenames("main", aspectFilename)
	prog, err := conf.Load()
	if err != nil {
		return nil, nil, err
	}
	initialPkgs := prog.InitialPackages()
	if len(initialPkgs) != 1 {
		return nil, nil, fmt.Errorf("unexpected initial packages: %v", initialPkgs)
	}
	pkgInfo := initialPkgs[0]
	if len(pkgInfo.Errors) != 0 {
		return nil, nil, fmt.Errorf("package %s has errors: %v", pkgInfo, pkgInfo.Errors)
	}
	if len(pkgInfo.Files) != 1 {
		return nil, nil, fmt.Errorf("only single aspect file is supported at the moment: %v", pkgInfo.Files)
	}
	return prog, pkgInfo, nil
}

func lookupAspects(pkg *types.Package, aspectIntf *types.Named) ([]*types.Named, error) {
	var result []*types.Named
	for _, name := range pkg.Scope().Names() {
		obj := pkg.Scope().Lookup(name)
		if fObj, ok := obj.(*types.Func); ok {
			if fObj.Name() == "main" {
				return nil, fmt.Errorf("main() is not supported in an aspect file: %s", fObj)
			}
		}
		if tObj, ok := obj.(*types.TypeName); ok {
			named := tObj.Type().(*types.Named)
			structureIsAspect := types.AssignableTo(named, aspectIntf)
			pointerIsAspect := types.AssignableTo(types.NewPointer(named), aspectIntf)
			if structureIsAspect {
				return nil, fmt.Errorf("aspect should have pointer-receiver: %s", named)
			}
			if pointerIsAspect {
				result = append(result, named)
			}
		}
	}
	return result, nil
}

func lookupAspectInterface(program *loader.Program) (*types.Named, error) {
	for pkg := range program.AllPackages {
		if pkg.Path() == aspectPackagePath {
			obj := pkg.Scope().Lookup("Aspect")
			tObj, ok := obj.(*types.TypeName)
			if !ok {
				return nil, fmt.Errorf("invalid aspect definition (not *types.TypeName)")
			}
			named, ok := tObj.Type().(*types.Named)
			if !ok {
				return nil, fmt.Errorf("invalid aspect definition (not *types.Named)")
			}
			if !types.IsInterface(named) {
				return nil, fmt.Errorf("invalid aspect definition (not interface)")
			}
			return named, nil
		}
	}
	return nil, fmt.Errorf("could not find %s", aspectPackagePath)
}
