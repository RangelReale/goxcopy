package goxcopy

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/RangelReale/rprim"
)

// The creator interface represent the target of a copy.
type XCopyCreator interface {
	// The type that will be created.
	Type() reflect.Type
	// Creates an instance of the value
	Create() (reflect.Value, error)
	// Sets the existing value.
	SetCurrentValue(current reflect.Value) error
	// Sets a field value.
	SetField(index reflect.Value, value reflect.Value) error
}

// Gets a creator for a type.
func (c *Config) XCopyGetCreator(ctx *Context, t reflect.Type) (XCopyCreator, error) {
	tkind := rprim.UnderliningTypeKind(t)

	switch tkind {
	case reflect.Struct:
		return &copyCreator_Struct{ctx: ctx, c: c, t: t}, nil
	case reflect.Map:
		return &copyCreator_Map{ctx: ctx, c: c, t: t}, nil
	case reflect.Slice, reflect.Array:
		return &copyCreator_Slice{ctx: ctx, c: c, t: t}, nil
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String, reflect.Interface:
		return &copyCreator_Primitive{ctx: ctx, c: c, t: t}, nil
	}
	return nil, newError(fmt.Errorf("Kind not supported: %s", tkind.String()), ctx.Dup())
}

//
// Struct
//

type copyCreator_Struct struct {
	ctx      *Context
	c        *Config
	t        reflect.Type
	isEnsure bool
	v        reflect.Value
}

func (c *copyCreator_Struct) Type() reflect.Type {
	return c.t
}

func (c *copyCreator_Struct) SetCurrentValue(current reflect.Value) error {
	if current.Type() != c.t {
		return newError(fmt.Errorf("Destination is not of the same type (%s -> %s)", current.Type().String(), c.t.String()), c.ctx.Dup())
	}
	if current.IsValid() {
		// check if must write on the passed value
		overwrite_existing := (c.c.Flags & XCF_OVERWRITE_EXISTING) == XCF_OVERWRITE_EXISTING
		need_duplicate := !overwrite_existing

		if overwrite_existing {
			if current.Kind() == reflect.Ptr && rprim.UnderliningValueIsNil(current) {
				// If is nil pointer, just set it, the value will be set later if the source is not nil
				c.v = current
			} else if rprim.UnderliningValue(current).NumField() == 0 || rprim.UnderliningValue(current).Field(0).CanSet() {
				// If the field is settable, set it as the value
				c.v = current
			} else {
				// fields are not settable, duplicate value if allowed
				need_duplicate = true
			}
		}

		if need_duplicate {
			if !overwrite_existing || (c.c.Flags&XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE) == XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE {
				// Create a new instance copying the value
				newValue, err := c.c.XCopyToNew(c.ctx, current, c.t)
				if err != nil {
					return err
				}
				c.v = newValue
			} else {
				return newError(fmt.Errorf("Struct fields are not settable and duplicates are not allowed"), c.ctx.Dup())
			}
		}
	}
	return nil
}

func (c *copyCreator_Struct) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *copyCreator_Struct) SetField(index reflect.Value, value reflect.Value) error {
	fieldname, err := c.c.RprimConfig.ConvertToString(index)
	if err != nil {
		return err
	}

	err = c.ensureValue()
	if err != nil {
		return err
	}

	ut := rprim.UnderliningType(c.t)
	uv := rprim.UnderliningValue(c.v)

	var fieldType reflect.StructField
	var fieldTypeOk bool

	for fi := 0; fi < ut.NumField(); fi++ {
		f := ut.Field(fi)
		fname := f.Name

		tag_fields := c.c.GetStructTagFields(f)
		if len(tag_fields) > 0 {
			if tag_fields[0] == "-" {
				fname = "" // skip
			} else {
				fname = tag_fields[0]
			}
		}

		if fname == fieldname {
			fieldType = f
			fieldTypeOk = true
			break
		}
	}
	if !fieldTypeOk {
		if (c.c.Flags & XCF_ERROR_IF_STRUCT_FIELD_MISSING) == XCF_ERROR_IF_STRUCT_FIELD_MISSING {
			return newError(fmt.Errorf("Field %s missing on struct", fieldname), c.ctx.Dup())
		}
		return nil
	}

	fieldValue := uv.FieldByName(fieldType.Name)
	if !fieldValue.CanSet() {
		return newError(fmt.Errorf("Struct field %s is not settable", fieldname), c.ctx.Dup())
	}

	cv, err := c.c.XCopyUsingExistingIfValid(c.ctx, value, fieldType.Type, fieldValue)
	if err != nil {
		return err
	}

	fieldValue.Set(cv)
	return nil
}

func (c *copyCreator_Struct) ensureValue() error {
	if !c.isEnsure {
		if !c.v.IsValid() {
			// if not valid, create a new instance
			c.v, _ = rprim.NewUnderliningValue(c.t)
		} else {
			// else ensure all pointer indirections are not nil
			_, err := rprim.EnsureUnderliningValue(c.v)
			if err != nil {
				return err
			}
		}
		c.isEnsure = true
	}
	return nil
}

func (c *copyCreator_Struct) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.New(c.t).Elem()
	}
}

//
// Map
//

type copyCreator_Map struct {
	ctx      *Context
	c        *Config
	t        reflect.Type
	isEnsure bool
	v        reflect.Value
}

func (c *copyCreator_Map) Type() reflect.Type {
	return c.t
}

func (c *copyCreator_Map) SetCurrentValue(current reflect.Value) error {
	if current.Type() != c.t {
		return newError(fmt.Errorf("Destination is not of the same type (%s -> %s)", current.Type().String(), c.t.String()), c.ctx.Dup())
	}

	if current.IsValid() {
		// check if must write on the passed value
		overwrite_existing := (c.c.Flags & XCF_OVERWRITE_EXISTING) == XCF_OVERWRITE_EXISTING
		need_duplicate := !overwrite_existing

		if overwrite_existing {
			if current.Kind() == reflect.Ptr || current.CanSet() {
				// If is nil pointer, just set it, the value will be set later if the source is not nil
				c.v = current
			} else {
				// fields are not settable, duplicate value if allowed
				need_duplicate = true
			}
		}

		if need_duplicate {
			if !overwrite_existing || (c.c.Flags&XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE) == XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE {
				// Create a new instance copying the value
				newValue, err := c.c.XCopyToNew(c.ctx, current, c.t)
				if err != nil {
					return err
				}
				c.v = newValue
			} else {
				return newError(fmt.Errorf("Slice is not settable and duplicates are not allowed"), c.ctx.Dup())
			}
		}
	}
	return nil
}

func (c *copyCreator_Map) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *copyCreator_Map) SetField(index reflect.Value, value reflect.Value) error {
	ut := rprim.UnderliningType(c.t)

	// convert index to the map index type
	mapindex, err := c.c.RprimConfig.Convert(index, ut.Key())
	if err != nil {
		return err
	}

	err = c.ensureValue()
	if err != nil {
		return err
	}

	uv := rprim.UnderliningValue(c.v)

	// map items are not settable, if set, must be copied
	currentValue := reflect.Value{}
	if cur := uv.MapIndex(mapindex); cur.IsValid() {
		if (c.c.Flags & XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE) == XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE {
			currentValue, err = c.c.XCopyToNew(c.ctx, cur, cur.Type())
			if err != nil {
				return err
			}
		} else {
			return newError(errors.New("Map item is not settable, cannot set existing item"), c.ctx.Dup())
		}
	}

	// special case of map[x]interface{} to allow inner maps of the same type as this
	target_type := ut.Elem()
	if !((c.c.Flags & XCF_DISABLE_MAPOFINTERFACE_TARGET_RECURSION) == XCF_DISABLE_MAPOFINTERFACE_TARGET_RECURSION) {
		if target_type.Kind() == reflect.Interface && KindHasFields(rprim.UnderliningValueKind(value)) {
			target_type = ut
		}
	}

	cv, err := c.c.XCopyUsingExistingIfValid(c.ctx, value, target_type, currentValue)
	if err != nil {
		return err
	}

	uv.SetMapIndex(mapindex, cv)
	return nil
}

func (c *copyCreator_Map) ensureValue() error {
	if !c.isEnsure {
		var last reflect.Value
		if !c.v.IsValid() {
			// if not valid, create a new instance
			c.v, last = rprim.NewUnderliningValue(c.t)
		} else {
			// else ensure all pointer indirections are not nil
			var err error
			last, err = rprim.EnsureUnderliningValue(c.v)
			if err != nil {
				return err
			}
		}
		if last.IsNil() {
			last.Set(reflect.MakeMap(rprim.UnderliningType(c.t)))
		}
		c.isEnsure = true
	}
	return nil
}

func (c *copyCreator_Map) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.New(c.t).Elem()
	}
}

//
// Slice
//

type copyCreator_Slice struct {
	ctx      *Context
	c        *Config
	t        reflect.Type
	isEnsure bool
	v        reflect.Value
}

func (c *copyCreator_Slice) Type() reflect.Type {
	return c.t
}

func (c *copyCreator_Slice) SetCurrentValue(current reflect.Value) error {
	if current.Type() != c.t {
		return newError(fmt.Errorf("Destination is not of the same type (%s -> %s)", current.Type().String(), c.t.String()), c.ctx.Dup())
	}

	if current.IsValid() {
		// check if must write on the passed value
		overwrite_existing := (c.c.Flags & XCF_OVERWRITE_EXISTING) == XCF_OVERWRITE_EXISTING
		need_duplicate := !overwrite_existing

		if overwrite_existing {
			if current.Kind() == reflect.Ptr || current.CanSet() {
				// If is nil pointer, just set it, the value will be set later if the source is not nil
				c.v = current
			} else {
				// fields are not settable, duplicate value if allowed
				need_duplicate = true
			}
		}

		if need_duplicate {
			if !overwrite_existing || (c.c.Flags&XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE) == XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE {
				// Create a new instance copying the value
				newValue, err := c.c.XCopyToNew(c.ctx, current, c.t)
				if err != nil {
					return err
				}
				c.v = newValue
			} else {
				return newError(fmt.Errorf("Slice is not settable and duplicates are not allowed"), c.ctx.Dup())
			}
		}
	}
	return nil
}

func (c *copyCreator_Slice) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *copyCreator_Slice) SetField(index reflect.Value, value reflect.Value) error {
	// convert index to int
	sliceindex, err := c.c.RprimConfig.Convert(index, reflect.TypeOf(0))
	if err != nil {
		return err
	}

	err = c.ensureValue()
	if err != nil {
		return err
	}

	ut := rprim.UnderliningType(c.t)
	uv := rprim.UnderliningValue(c.v)

	// Add zero values until the index
	for int(sliceindex.Int()) >= uv.Len() {
		c.append()
	}

	currentValue := reflect.Value{}
	if int(sliceindex.Int()) < uv.Len() {
		currentValue = uv.Index(int(sliceindex.Int()))
	}

	cv, err := c.c.XCopyUsingExistingIfValid(c.ctx, value, ut.Elem(), currentValue)
	if err != nil {
		return err
	}

	uv.Index(int(sliceindex.Int())).Set(cv)
	return nil
}

func (c *copyCreator_Slice) append() {
	v := rprim.UnderliningValue(c.v)
	if v.CanSet() {
		v.Set(reflect.Append(v, reflect.Zero(rprim.UnderliningType(c.t).Elem())))
	} else {
		panic("Should not happen")
	}
}

func (c *copyCreator_Slice) ensureValue() error {
	if !c.isEnsure {
		var last reflect.Value
		if !c.v.IsValid() {
			// if not valid, create a new instance
			c.v, last = rprim.NewUnderliningValue(c.t)
		} else {
			// else ensure all pointer indirections are not nil
			var err error
			last, err = rprim.EnsureUnderliningValue(c.v)
			if err != nil {
				return err
			}
		}
		if last.Kind() != reflect.Array {
			if last.IsNil() {
				last.Set(reflect.MakeSlice(rprim.UnderliningType(c.t), 0, 0))
			}
		}
		c.isEnsure = true
	}
	return nil
}

func (c *copyCreator_Slice) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.New(c.t).Elem()
	}
}

//
// Primitive
//

type copyCreator_Primitive struct {
	ctx *Context
	c   *Config
	t   reflect.Type
	//it  reflect.Type
	isEnsure bool
	v        reflect.Value
}

func (c *copyCreator_Primitive) Type() reflect.Type {
	return c.t
}

func (c *copyCreator_Primitive) SetCurrentValue(current reflect.Value) error {
	if current.Type() != c.t {
		return newError(fmt.Errorf("Destination is not of the same type (%s -> %s)", current.Type().String(), c.t.String()), c.ctx.Dup())
	}

	if current.IsValid() {
		// check if must write on the passed value
		overwrite_existing := (c.c.Flags & XCF_OVERWRITE_EXISTING) == XCF_OVERWRITE_EXISTING
		need_duplicate := !overwrite_existing

		if overwrite_existing {
			if current.Kind() == reflect.Ptr || current.CanSet() {
				// If is nil pointer, just set it, the value will be set later if the source is not nil
				c.v = current
			} else {
				// fields are not settable, duplicate value if allowed
				need_duplicate = true
			}
		}

		if need_duplicate {
			if !overwrite_existing || (c.c.Flags&XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE) == XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE {
				// Create a new instance copying the value
				newValue, err := c.c.XCopyToNew(c.ctx, current, c.t)
				if err != nil {
					return err
				}
				c.v = newValue
			} else {
				return newError(fmt.Errorf("Slice is not settable and duplicates are not allowed"), c.ctx.Dup())
			}
		}
	}
	return nil
}

func (c *copyCreator_Primitive) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *copyCreator_Primitive) SetField(index reflect.Value, value reflect.Value) error {
	if index.IsValid() {
		return newError(fmt.Errorf("Cannot set a primitive with an index"), c.ctx.Dup())
	}

	err := c.ensureValue()
	if err != nil {
		return err
	}

	val, err := c.c.RprimConfig.Convert(value, c.t)
	if err != nil {
		return err
	}

	// check if settable
	if c.v.CanSet() {
		c.v.Set(val)
	} else if c.v.Kind() == reflect.Ptr && val.Kind() == reflect.Ptr {
		// if is non-nil pointer, set the pointed to element value
		if c.v.IsNil() && !val.IsNil() {
			return newError(errors.New("The primitive value is not settable, and the destination value is nil"), c.ctx.Dup())
		} else if val.IsNil() {
			return newError(errors.New("The primitive value is not settable, and the source value is nil"), c.ctx.Dup())
		} else {
			c.v.Elem().Set(val.Elem())
		}
	} else {
		return newError(errors.New("The primitive value is not settable"), c.ctx.Dup())
	}
	return nil
}

func (c *copyCreator_Primitive) ensureValue() error {
	if !c.isEnsure {
		if !c.v.IsValid() {
			// if not valid, create a new instance
			c.v, _ = rprim.NewUnderliningValue(c.t)
		} else {
			// else ensure all pointer indirections are not nil
			_, err := rprim.EnsureUnderliningValue(c.v)
			if err != nil {
				return err
			}
		}
		c.isEnsure = true
	}
	return nil
}

func (c *copyCreator_Primitive) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.New(c.t).Elem()
	}
}
