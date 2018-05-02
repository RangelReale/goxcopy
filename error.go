package goxcopy

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
	return e.Err.Error()
}
