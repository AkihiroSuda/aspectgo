package main

import (
	"fmt"
	"regexp"

	asp "golang.org/x/exp/aspectgo/aspect"
)

// Aspect1 is ineffective because it is currently overrided by Aspect2.
// Note that future implementation may make Aspect1 effective.
type Aspect1 struct {
}

func (a *Aspect1) Pointcut() asp.Pointcut {
	pkg := regexp.QuoteMeta("golang.org/x/exp/aspectgo/example/multipointcut")
	s := regexp.QuoteMeta(pkg + ".sayHello")
	return asp.NewCallPointcutFromRegexp(s)
}

func (a *Aspect1) Advice(ctx asp.Context) []interface{} {
	args := ctx.Args()
	fmt.Println("Aspect1")
	res := ctx.Call(args)
	return res
}

// Aspect2 is effective.
type Aspect2 struct {
}

func (a *Aspect2) Pointcut() asp.Pointcut {
	pkg := regexp.QuoteMeta("golang.org/x/exp/aspectgo/example/multipointcut")
	s := pkg + ".*"
	return asp.NewCallPointcutFromRegexp(s)
}

func (a *Aspect2) Advice(ctx asp.Context) []interface{} {
	args := ctx.Args()
	fmt.Println("Aspect2")
	res := ctx.Call(args)
	return res
}
