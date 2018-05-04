package goxcopy

import (
	"fmt"
	"io"
	"reflect"
)

type Callback interface {
	BeginNew(ctx *Context, src reflect.Value, destType reflect.Type)
	EndNew(ctx *Context, src reflect.Value, destType reflect.Type)

	// Call just before a field is pushed
	PushField(ctx *Context, fieldname reflect.Value, src reflect.Value, dest Creator)
	// Call just before a field is popped
	PopField(ctx *Context, fieldname reflect.Value, src reflect.Value, dest Creator)

	BeforeSetValue(ctx *Context, src reflect.Value, dest Creator, value reflect.Value)
	AfterSetValue(ctx *Context, src reflect.Value, dest Creator, value reflect.Value)
}

// Debug callback

type DebugCallback struct {
	W io.Writer
}

func NewDebugCallback(w io.Writer) *DebugCallback {
	return &DebugCallback{
		W: w,
	}
}

func (c *DebugCallback) BeginNew(ctx *Context, src reflect.Value, destType reflect.Type) {
	fmt.Fprintf(c.W, ">>> BEGIN NEW: %s -> [%s]\n", destType.Kind(), ctx.FieldsAsString())
}

func (c *DebugCallback) EndNew(ctx *Context, src reflect.Value, destType reflect.Type) {
	fmt.Fprintf(c.W, "<<< END NEW: %s -> [%s]\n", destType.Kind(), ctx.FieldsAsString())
}

func (c *DebugCallback) PushField(ctx *Context, fieldname reflect.Value, src reflect.Value, dest Creator) {
	fmt.Fprintf(c.W, "+++ PUSH FIELD: %s -> [%s]\n", FieldnameToString(fieldname), ctx.FieldsAsString())
}

func (c *DebugCallback) PopField(ctx *Context, fieldname reflect.Value, src reflect.Value, dest Creator) {
	fmt.Fprintf(c.W, "--- POP FIELD: %s -> [%s]\n", FieldnameToString(fieldname), ctx.FieldsAsString())
}

func (c *DebugCallback) BeforeSetValue(ctx *Context, src reflect.Value, dest Creator, value reflect.Value) {
}

func (c *DebugCallback) AfterSetValue(ctx *Context, src reflect.Value, dest Creator, value reflect.Value) {
	fmt.Fprintf(c.W, "=== SET VALUE: %s [%s]\n", ctx.FieldsAsString(), dest.Type().Kind().String())
}
