package main

import (
	"fmt"
	"regexp"

	asp "golang.org/x/exp/aspectgo/aspect"
)

// SAspect won't be woven, because it's not an "execution" pointcut
type SAspect struct {
}

func (a *SAspect) Pointcut() asp.Pointcut {
	s := regexp.QuoteMeta("(*golang.org/x/exp/aspectgo/example/receiver.S).Foo")
	return asp.NewCallPointcutFromRegexp(s)
}
func (a *SAspect) Advice(ctx asp.Context) []interface{} {
	return advice("SAspect", ctx)
}

// IAspect will be woven
type IAspect struct {
}

func (a *IAspect) Pointcut() asp.Pointcut {
	s := regexp.QuoteMeta("(golang.org/x/exp/aspectgo/example/receiver.I).Foo")
	return asp.NewCallPointcutFromRegexp(s)
}
func (a *IAspect) Advice(ctx asp.Context) []interface{} {
	return advice("IAspect", ctx)
}

func advice(name string, ctx asp.Context) []interface{} {
	args, recv := ctx.Args(), ctx.Receiver()
	fmt.Printf("%s BEFORE call (args=%+v, recv=%+v)\n",
		name, args, recv)
	res := ctx.Call(args)
	fmt.Printf("%s AFTER call (res=%+v)\n",
		name, res)
	return res
}
