// Package aspect provides interfaces for AspectGo AOP framework.
package aspect

import (
	"fmt"
)

// Context is the type for joinpoint context definition.
type Context interface {
	// Args returns the original argument set for the joinpoint.
	// The slice can be empty []interface{}{}, but cannot be nil.
	Args() []interface{}

	// Call calls the joinpoint.
	// User must be careful about the length and the type of
	// the []interface{} slices.
	// The slices can be empty []interface{}{}, but cannot be nil.
	Call([]interface{}) []interface{}

	// Receiver returns the receiver for methods.
	// For non-method function, it just returns nil.
	Receiver() interface{}
}

// Pointcut is the type for pointcut definition.
// User should NOT be aware of the internal representation. (string)
// Currently, only "call" pointcut is supported.
// TODO: support "execution" pointcut.
type Pointcut string

func (pc Pointcut) String() string {
	return string(pc)
}

// NewCallPointcutFromRegexp creates a "call" pointcut from s.
// s needs to be a regexp for function/method name.
func NewCallPointcutFromRegexp(s string) Pointcut {
	return Pointcut(s)
}

// NewExecPointcutFromRegexp creates a "execution" pointcut from s.
// s needs to be a regexp for function/method name.
func NewExecPointcutFromRegexp(s string) Pointcut {
	panic(fmt.Errorf("\"execution\" pointcut is not implemented yet"))
}

// Aspect is the interface for aspect definition.
// Currently, only "around" type advice is supported.
type Aspect interface {
	// Pointcut returns the pointcut for the aspect.
	// Pointcut is executed on compilation-time.
	Pointcut() Pointcut

	// Advice executes the "around" advice.
	// User must be careful about the length and the type of
	// the []interface{} slice.
	// The slice can be empty []interface{}{}, but cannot be nil.
	Advice(Context) []interface{}
}
