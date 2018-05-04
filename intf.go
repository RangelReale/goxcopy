package goxcopy

import (
	"fmt"
	"io"
	"reflect"
	"strings"
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
	W     io.Writer
	level int
}

func NewDebugCallback(w io.Writer) *DebugCallback {
	return &DebugCallback{
		W: w,
	}
}

func (c *DebugCallback) BeginNew(ctx *Context, src reflect.Value, destType reflect.Type) {
	ts := strings.Repeat("\t", c.level+1)
	c.level++
	fmt.Fprintf(c.W, "%s>>> BEGIN NEW: %s -> [%s]\n", ts, destType.Kind(), ctx.FieldsAsString())
}

func (c *DebugCallback) EndNew(ctx *Context, src reflect.Value, destType reflect.Type) {
	c.level--
	ts := strings.Repeat("\t", c.level+1)
	fmt.Fprintf(c.W, "%s<<< END NEW: %s -> [%s]\n", ts, destType.Kind(), ctx.FieldsAsString())
}

func (c *DebugCallback) PushField(ctx *Context, fieldname reflect.Value, src reflect.Value, dest Creator) {
	ts := strings.Repeat("\t", c.level+1)
	c.level++
	fmt.Fprintf(c.W, "%s+++ PUSH FIELD: %s -> [%s]\n", ts, FieldnameToString(fieldname), ctx.FieldsAsString())
}

func (c *DebugCallback) PopField(ctx *Context, fieldname reflect.Value, src reflect.Value, dest Creator) {
	c.level--
	ts := strings.Repeat("\t", c.level+1)
	fmt.Fprintf(c.W, "%s--- POP FIELD: %s -> [%s]\n", ts, FieldnameToString(fieldname), ctx.FieldsAsString())
}

func (c *DebugCallback) BeforeSetValue(ctx *Context, src reflect.Value, dest Creator, value reflect.Value) {
}

func (c *DebugCallback) AfterSetValue(ctx *Context, src reflect.Value, dest Creator, value reflect.Value) {
	ts := strings.Repeat("\t", c.level+1)
	fmt.Fprintf(c.W, "%s=== SET VALUE: %s [%s]\n", ts, ctx.FieldsAsString(), dest.Type().Kind().String())
}
