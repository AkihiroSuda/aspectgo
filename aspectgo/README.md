# AspectGo: AOP Framework for Go

AspectGo is an AOP framework for Go.
You can write an aspect in a simple Go-compatible grammar.

I'm planning to propose merging AspectGo to the upstream of [`golang.org/x/exp`](https://godoc.org/golang.org/x/exp) if possible.

Recipe:

 * Logging
 * Assertion
 * Fault injection
 * Mocking
 * Coverage-guided genetic fuzzing (as in [AFL](http://lcamtuf.coredump.cx/afl/technical_details.txt))
 * Fuzzed(randomized) scheduling

The interface is not fixed yet.
Your suggestion and PR are welcome.

## Install

    $ git clone https://github.com/AkihiroSuda/golang-exp-aspectgo.git
    $ ln -s $(pwd)/golang-exp-aspectgo $GOPATH/src/golang.org/x/exp
    $ go build golang.org/x/exp/aspectgo/cmd/aspectgo

## Example

    $ go build golang.org/x/exp/aspectgo/example/hello && ./hello
    hello
    $ aspectgo \
      -w /tmp/wovengopath \                         # output gopath
      -t golang.org/x/exp/aspectgo/example/hello \  # target package
      example/hello/main_aspect.go                  # aspect file
    $ GOPATH=/tmp/wovengopath go build golang.org/x/exp/aspectgo/example/hello && ./hello
    BEFORE hello
    hello
    AFTER hello

The aspect is located on [example/hello/main_aspect.go](example/hello/main_aspect.go):

```go
package main

import (
	"fmt"
	"regexp"

	asp "golang.org/x/exp/aspectgo/aspect"
)

// ExampleAspect implements interface asp.Aspect
type ExampleAspect struct {
}

// Executed on compilation-time
func (a *ExampleAspect) Pointcut() asp.Pointcut {
	pkg := regexp.QuoteMeta("golang.org/x/exp/aspectgo/example/hello")
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
```


The target is [example/hello/main.go](example/hello/main.go):
```go
package main

import (
	"fmt"
)

func sayHello(s string) {
	fmt.Println("hello " + s)
}

func main() {
	sayHello("world")
}
```


You can also execute other examples as follows:

    $ go test -v golang.org/x/exp/aspectgo/example

If the output is hard to read, please add the `-parallel 1` flag to `go test`.

## Hint

 * Clean `/tmp/wovengopath` before running `aspectgo` every time.
 * Clean GOPATH before running `aspectgo` for faster compilation.

## Current Limitation

 * Only single aspect file is supported (But you can define multiple aspects in a single file)
 * Only regexp for function name (excluding `main` and `init`) and method name can be a pointcut
 * Only "call" pointcut is supported. No support for "execution" pointcut yet:
   * Suppose that `*S`, `*T` implements `I`, and there is a call to `I.Foo()` in the target package. You can make a pointcut for `I.Foo()`, but you can't make a pointcut for `*S` nor `*T`.
   * Aspect cannot be woven to Go-builtin packages. i.e., You can't hook a call _from_ a Go-builtin pacakge. (But you can hook a call _to_ a Go-builtin package by just making a "call" pointcut for it)
 * Only "around" advice is supported. No support for "before" and "after" pointcut.
 * If an object hits multiple pointcuts, only the last one is effective.
 
## Related Work

 * [github.com/deferpanic/goweave](https://github.com/deferpanic/goweave)
 * [github.com/gogap/aop](https://github.com/gogap/aop)
 * [golang.org/x/tools/refactor/eg](https://github.com/golang/tools/blob/master/refactor/eg/eg.go) (Not AOP, but has some similarity)
 * [github.com/coreos/gofail](https://github.com/coreos/gofail) (ditto)
