package main

import (
	"fmt"
	"os"
	"regexp"

	asp "golang.org/x/exp/aspectgo/aspect"
)

// ExampleAspect implements interface asp.Aspect
type ExampleAspect struct {
}

// Executed on compilation-time
func (a *ExampleAspect) Pointcut() asp.Pointcut {
	pkg := regexp.QuoteMeta("golang.org/x/exp/aspectgo/example/hello2")
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

// FmtPrintlnAspect implements interface asp.Aspect
type FmtPrintlnAspect struct {
}

func (a *FmtPrintlnAspect) Pointcut() asp.Pointcut {
	s := regexp.QuoteMeta("fmt.Println")
	return asp.NewCallPointcutFromRegexp(s)
}

func (a *FmtPrintlnAspect) Advice(ctx asp.Context) []interface{} {
	args := ctx.Args()
	fmt.Fprintf(os.Stderr, "directing to stderr: %s\n", args...)
	return []interface{}{0, nil}
}
