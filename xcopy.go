package goxcopy

import (
	"errors"
	"reflect"
)

// Copy a source variable to a destination variable, overwriting it.
// The destination variable must be settable.
// This is an alias for "CopyToExisting"
// The src variable is never changed in any circunstance.
func Copy(from interface{}, to interface{}) error {
	return CopyToExisting(from, to)
}

// Copy a source variable to a new instance of the passed type.
// The src variable is never changed in any circunstance.
func CopyToNew(src interface{}, destType reflect.Type) (interface{}, error) {
	return NewConfig().CopyToNew(src, destType)
}

// Copy a source variable to a new instance of the type of the passed value.
// The passed variable is used to initialize the new instance value, but is not
// changed in any way.
// The src and currentValue variable are never changed in any circunstance.
func CopyUsingExisting(src interface{}, currentValue interface{}) (interface{}, error) {
	return NewConfig().CopyUsingExisting(src, currentValue)
}

// Copy a source variable to a destination variable, overwriting it.
// The destination variable must be settable.
// This is an alias for "CopyToExisting"
// The src variable is never changed in any circunstance.
func CopyToExisting(src interface{}, currentValue interface{}) error {
	return NewConfig().CopyToExisting(src, currentValue)
}

// Merges all source variables to a new instance of the passed type.
// The src variables are never changed in any circunstance.
func MergeToNew(destType reflect.Type, src ...interface{}) (interface{}, error) {
	return NewConfig().MergeToNew(destType, src...)
}

// Copy a source variable to a new instance of the passed type.
// The src variable is never changed in any circunstance.
func XCopyToNew(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	return NewConfig().XCopyToNew(NewContext(), src, destType)
}

// Copy a source variable to a new instance of the type of the passed value.
// The passed variable is used to initialize the new instance value, but is not
// changed in any way.
// The src and currentValue variable are never changed in any circunstance.
func XCopyUsingExisting(src reflect.Value, currentValue reflect.Value) (reflect.Value, error) {
	return NewConfig().XCopyUsingExisting(NewContext(), src, currentValue)
}

// Copy a source variable to a destination variable, overwriting it.
// The destination variable must be settable.
// This is an alias for "CopyToExisting"
// The src variable is never changed in any circunstance.
func XCopyToExisting(src reflect.Value, currentValue reflect.Value) error {
	return NewConfig().XCopyToExisting(NewContext(), src, currentValue)
}

// Merges all source variables to a new instance of the passed type.
// The src variables are never changed in any circunstance.
func XMergeToNew(destType reflect.Type, src ...reflect.Value) (reflect.Value, error) {
	return NewConfig().XMergeToNew(NewContext(), destType, src...)
}

// Copy a source variable to a new instance of the passed type.
// The src variable is never changed in any circunstance.
func (c *Config) CopyToNew(src interface{}, destType reflect.Type) (interface{}, error) {
	ret, err := c.XCopyToNew(NewContext(), reflect.ValueOf(src), destType)
	if err != nil {
		return nil, err
	}
	return ret.Interface(), nil
}

// Copy a source variable to a new instance of the type of the passed value.
// The passed variable is used to initialize the new instance value, but is not
// changed in any way.
// The src and currentValue variable are never changed in any circunstance.
func (c *Config) CopyUsingExisting(src interface{}, currentValue interface{}) (interface{}, error) {
	ret, err := c.XCopyUsingExisting(NewContext(), reflect.ValueOf(src), reflect.ValueOf(currentValue))
	if err != nil {
		return nil, err
	}
	return ret.Interface(), nil
}

// Copy a source variable to a destination variable, overwriting it.
// The destination variable must be settable.
// This is an alias for "CopyToExisting"
// The src variable is never changed in any circunstance.
func (c *Config) CopyToExisting(src interface{}, currentValue interface{}) error {
	return c.XCopyToExisting(NewContext(), reflect.ValueOf(src), reflect.ValueOf(currentValue))
}

// Merges all source variables to a new instance of the passed type.
// The src variables are never changed in any circunstance.
func (c *Config) MergeToNew(destType reflect.Type, src ...interface{}) (interface{}, error) {
	var rsrc []reflect.Value
	for _, isrc := range src {
		rsrc = append(rsrc, reflect.ValueOf(isrc))
	}

	ret, err := c.XMergeToNew(NewContext(), destType, rsrc...)
	if err != nil {
		return nil, err
	}
	return ret.Interface(), nil
}

// Copy a source variable to a new instance of the passed type.
// The src variable is never changed in any circunstance.
func (c *Config) XCopyToNew(ctx *Context, src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	c.callbackBeginNew(ctx, src, destType) // callback
	ret, err := c.internalXCopyUsingExistingIfValid(ctx, src, destType, reflect.Value{})
	c.callbackEndNew(ctx, src, destType) // callback
	return ret, err
}

// Copy a source variable to a new instance of the type of the passed value.
// The passed variable is used to initialize the new instance value, but is not
// changed in any way.
// The src and currentValue variable are never changed in any circunstance.
func (c *Config) XCopyUsingExisting(ctx *Context, src reflect.Value, currentValue reflect.Value) (reflect.Value, error) {
	return c.internalXCopyUsingExistingIfValid(ctx, src, reflect.TypeOf(currentValue.Interface()), currentValue)
}

// Copy a source variable to a destination variable, overwriting it.
// The destination variable must be settable.
// This is an alias for "CopyToExisting"
// The src variable is never changed in any circunstance.
func (c *Config) XCopyToExisting(ctx *Context, src reflect.Value, currentValue reflect.Value) error {
	_, err := c.Dup().AddFlags(XCF_OVERWRITE_EXISTING).internalXCopyUsingExistingIfValid(ctx, src, reflect.TypeOf(currentValue.Interface()), currentValue)
	return err
}

// Merges all source variables to a new instance of the passed type.
// The src variables are never changed in any circunstance.
func (c *Config) XMergeToNew(ctx *Context, destType reflect.Type, src ...reflect.Value) (reflect.Value, error) {
	if len(src) == 0 {
		return reflect.Value{}, newError(errors.New("At least one source is needed for merge"), ctx)
	}
	ret := reflect.Value{}
	var err error
	for _, isrc := range src {
		if !ret.IsValid() {
			// the first one must be created
			ret, err = c.XCopyToNew(ctx, isrc, destType)
			if err != nil {
				return reflect.Value{}, err
			}
			if !ret.CanAddr() {
				return reflect.Value{}, newError(errors.New("Destination value is not addressable, cannot continue merge"), ctx)
			}
			ret = ret.Addr()
		} else {
			// merge the rest
			err = c.XCopyToExisting(ctx, isrc, ret)
			if err != nil {
				return reflect.Value{}, err
			}
		}
	}
	return ret.Elem(), nil
}
