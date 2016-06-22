package main

import (
	"fmt"
	"regexp"

	asp "github.com/AkihiroSuda/aspectgo/aspect"
)

// ExampleAspect implements interface asp.Aspect
type ExampleAspect struct {
}

// Executed on compilation-time
func (a *ExampleAspect) Pointcut() asp.Pointcut {
	pkg := regexp.QuoteMeta("github.com/AkihiroSuda/aspectgo/example/hello")
	s := pkg + ".*"
	return asp.NewCallPointcutFromRegexp(s)
}

// Executed ONLY on runtime
func (a *ExampleAspect) Advice(ctx asp.Context) []interface{} {
	args := ctx.Args()
	fmt.Println("BEFORE hello")
	res := ctx.Call(args)
	fmt.Println("AFTER hello")
	return res
}
