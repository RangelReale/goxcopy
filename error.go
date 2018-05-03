package goxcopy

import "fmt"

// Error type
type Error struct {
	// Underlining error
	Err error
	// The context where the error occured
	Ctx *Context
}

func newError(err error, ctx *Context) *Error {
	return &Error{
		Err: err,
		Ctx: ctx,
	}
}

func (e *Error) Error() string {
	if len(e.Ctx.Fields) > 0 {
		return fmt.Sprintf("%s [%s]", e.Err.Error(), e.Ctx.FieldsAsString())
	} else {
		return e.Err.Error()
	}
}
