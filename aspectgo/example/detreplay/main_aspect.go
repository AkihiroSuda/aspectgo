package main

import (
	"fmt"
	"time"

	asp "golang.org/x/exp/aspectgo/aspect"

	"golang.org/x/exp/aspectgo/example/detreplay/worker"
)

type DetAspect struct {
}

func (a *DetAspect) Pointcut() asp.Pointcut {
	s := ".*"
	return asp.NewCallPointcutFromRegexp(s)
}
func (a *DetAspect) Advice(ctx asp.Context) []interface{} {
	args := ctx.Args()
	recv, ok := ctx.Receiver().(*worker.W)
	if ok {
		// this sleep increases determinism!
		time.Sleep(time.Duration(recv.X*10) * time.Millisecond)
		fmt.Printf("hook %v\n", recv)
	}
	res := ctx.Call(args)
	return res
}
