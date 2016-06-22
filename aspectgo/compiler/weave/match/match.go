package match

import (
	"go/ast"
	"go/types"
	"log"
	"regexp"

	"golang.org/x/tools/go/loader"

	"golang.org/x/exp/aspectgo/aspect"
	"golang.org/x/exp/aspectgo/compiler/util"
)

// ObjMatchPointcut returns true if obj matches the pointcut.
// current implementation is very naive: just checks regexp for types.Func.FullName()
// TODO: support interface pointcut
func ObjMatchPointcut(prog *loader.Program, id *ast.Ident, obj types.Object, pointcut aspect.Pointcut) bool {
	fn, ok := obj.(*types.Func)
	if ok {
		return fnObjMatchPointcutByRegexp(fn, pointcut)
	}
	return false
}

func fnObjMatchPointcutByRegexp(fn *types.Func, pointcut aspect.Pointcut) bool {
	// TODO: cache compiled regexp
	re, err := regexp.Compile(string(pointcut))
	if err != nil {
		log.Printf("pointcut %s is not a valid regexp: %s", pointcut, err)
		return false
	}
	matched := re.MatchString(fn.FullName())
	if util.DebugMode {
		log.Printf("matched=%t for %s (pointcut=%s)", matched, fn.FullName(), string(pointcut))
	}
	return matched
}
