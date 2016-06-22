package main

import (
	"fmt"
	"regexp"

	asp "github.com/AkihiroSuda/aspectgo/aspect"
)

type Pkg1SAspect struct {
}

func (a *Pkg1SAspect) Pointcut() asp.Pointcut {
	s := regexp.QuoteMeta("(*github.com/AkihiroSuda/aspectgo/example/receiver2/pkg1.S).Foo")
	return asp.NewCallPointcutFromRegexp(s)
}
func (a *Pkg1SAspect) Advice(ctx asp.Context) []interface{} {
	args, recv := ctx.Args(), ctx.Receiver()
	fmt.Printf("Pkg1SAspect hook (args=%+v, recv=%+v)\n",
		args, recv)
	res := ctx.Call(args)
	return res
}
