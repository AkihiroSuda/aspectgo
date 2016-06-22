package main

import (
	asp "golang.org/x/exp/aspectgo/aspect"
)

// EverythingAspect implements interface asp.Aspect.
// This aspect is useful for smoke-testing arbitrary target:
// aspectgo -t github.com/foo/bar/... example/everything_aspect.go
type EverythingAspect struct {
}

// Executed on compilation-time
func (a *EverythingAspect) Pointcut() asp.Pointcut {
	return asp.NewCallPointcutFromRegexp(".*")
}

// Executed ONLY on runtime
func (a *EverythingAspect) Advice(ctx asp.Context) []interface{} {
	// NOP
	return ctx.Call(ctx.Args())
}
