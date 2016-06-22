package rt

import (
	"fmt"
	"runtime/debug"
	"testing"

	asp "golang.org/x/exp/aspectgo/aspect"
)

type dummyAspect struct {
}

func (a *dummyAspect) Pointcut() asp.Pointcut {
	return asp.Pointcut("dummy")
}

func (a *dummyAspect) Advice(ctx asp.Context) []interface{} {
	args := ctx.Args()
	fmt.Println("BEFORE hello")
	debug.PrintStack()
	res := ctx.Call(args)
	fmt.Println("AFTER hello")
	return res
}

func sayHello(s string) {
	fmt.Println("hello " + s)
}

func TestContextImpl(t *testing.T) {
	// woven expression for `sayHello("world")`
	(&dummyAspect{}).Advice(
		&ContextImpl{
			XArgs: []interface{}{"world"},
			XFunc: func(_ag_args []interface{}) []interface{} {
				_ag_arg0 := _ag_args[0].(string)
				sayHello(_ag_arg0)
				_ag_res := []interface{}{}
				return _ag_res
			}})
}
