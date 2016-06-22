package main

import (
	"fmt"

	asp "golang.org/x/exp/aspectgo/aspect"
)

// ExampleAspect implements interface asp.Aspect
type ExampleAspect struct {
}

// Executed on compilation-time
func (a *ExampleAspect) Pointcut() asp.Pointcut {
	s := "fmt\\.Print.*"
	return asp.NewCallPointcutFromRegexp(s)
}

// Executed ONLY on runtime
func (a *ExampleAspect) Advice(ctx asp.Context) []interface{} {
	args := ctx.Args()
	fmt.Printf("Hooking (args=%v)\n", args)
	res := ctx.Call(args)
	return res
}
