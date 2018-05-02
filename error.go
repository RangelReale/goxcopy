package goxcopy

type Error struct {
	Err error
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
