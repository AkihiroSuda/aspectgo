# AspectGo: AOP Framework for Go

[![Join the chat at https://gitter.im/AkihiroSuda/aspectgo](https://img.shields.io/badge/GITTER-join%20chat-green.svg)](https://gitter.im/AkihiroSuda/aspectgo?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Build Status](https://travis-ci.org/AkihiroSuda/aspectgo.svg?branch=master)](https://travis-ci.org/AkihiroSuda/aspectgo)
[![Go Report Card](https://goreportcard.com/badge/github.com/AkihiroSuda/aspectgo)](https://goreportcard.com/report/github.com/AkihiroSuda/aspectgo)

AspectGo is an AOP framework for Go.
You can write an aspect in a simple Go-compatible grammar.

I'm hoping to propose merging AspectGo to the upstream of [`golang.org/x/exp`](https://godoc.org/golang.org/x/exp) if possible, but no concret plan yet.

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

    $ go install github.com/AkihiroSuda/aspectgo/cmd/aspectgo

## Example

    $ go build github.com/AkihiroSuda/aspectgo/example/hello && ./hello
    hello
    $ aspectgo \
      -w /tmp/wovengopath \                         # output gopath
      -t github.com/AkihiroSuda/aspectgo/example/hello \  # target package
      example/hello/main_aspect.go                  # aspect file
    $ GOPATH=/tmp/wovengopath go build github.com/AkihiroSuda/aspectgo/example/hello && ./hello
    BEFORE hello
    hello
    AFTER hello

The aspect is located on [example/hello/main_aspect.go](example/hello/main_aspect.go):

```go
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

    $ go test -v github.com/AkihiroSuda/aspectgo/example

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
