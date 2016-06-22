// Package rt provides the internal runtime package for AspectGo.
// Do NOT access rt from an aspect file.
package rt

// ContextImpl implements aspect.Context
type ContextImpl struct {
	// XArgs should NOT be accessed manually.
	XArgs []interface{}

	// XFunc should NOT be accessed manually.
	XFunc func(args []interface{}) []interface{}

	// XReceiver should NOT be accessed manually.
	XReceiver interface{}
}

// Args should NOT be called manually.
func (ctx *ContextImpl) Args() []interface{} {
	return ctx.XArgs
}

// Call should NOT be called manually.
func (ctx *ContextImpl) Call(args []interface{}) []interface{} {
	return ctx.XFunc(args)
}

// Receiver should NOT be called manually.
func (ctx *ContextImpl) Receiver() interface{} {
	return ctx.XReceiver
}
