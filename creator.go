package goxcopy

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/RangelReale/rprim"
)

type XCopyCreator interface {
	Type() reflect.Type
	Create() (reflect.Value, error)
	SetCurrentValue(current reflect.Value) error
	SetField(index reflect.Value, value reflect.Value) error
}

func (c *Config) XCopyGetCreator(t reflect.Type) (XCopyCreator, error) {
	switch rprim.IndirectType(t).Kind() {
	case reflect.Struct:
		return &copyCreator_Struct{c: c, t: t}, nil
	case reflect.Map:
		return &copyCreator_Map{c: c, t: t}, nil
	case reflect.Slice:
		return &copyCreator_Slice{c: c, t: t}, nil
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String, reflect.Interface:
		return &copyCreator_Primitive{c: c, t: t}, nil
	}
	return nil, fmt.Errorf("Kind not supported: %s", t.Kind().String())
}

//
// Struct
//

type copyCreator_Struct struct {
	c  *Config
	t  reflect.Type
	it reflect.Type
	v  reflect.Value
}

func (c *copyCreator_Struct) Type() reflect.Type {
	return c.t
}

func (c *copyCreator_Struct) SetCurrentValue(current reflect.Value) error {
	if current.Type() != c.t {
		return fmt.Errorf("Destination is not of the same type (%s -> %s)", current.Type().String(), c.t.String())
	}
	if current.Kind() != reflect.Ptr || !current.IsNil() {
		forceduplicate := !((c.c.Flags & XCF_OVERWRITE_EXISTING) == XCF_OVERWRITE_EXISTING)

		// Check if the first field is settable
		if !forceduplicate && (reflect.Indirect(current).NumField() == 0 || reflect.Indirect(current).Field(0).CanSet()) {
			c.v = current
		} else {
			if forceduplicate || (c.c.Flags&XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE) == XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE {
				// if struct is not settable, must make a copy
				newValue, err := c.c.XCopyToNew(current, c.t)
				if err != nil {
					return err
				}
				c.v = newValue
			} else {
				return fmt.Errorf("Struct field is not settable")
			}
		}
		c.it = rprim.IndirectType(c.t)
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

	c.ensureValue()

	var fieldType reflect.StructField
	var fieldTypeOk bool

	for fi := 0; fi < c.it.NumField(); fi++ {
		f := c.it.Field(fi)
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
			return fmt.Errorf("Field %s missing on struct", fieldname)
		}
		return nil
	}

	fieldValue := reflect.Indirect(c.v).FieldByName(fieldType.Name)
	if !fieldValue.CanSet() {
		return fmt.Errorf("Struct field %s is not settable", fieldname)
	}

	cv, err := c.c.XCopyUsingExistingIfValid(value, fieldType.Type, fieldValue)
	if err != nil {
		return err
	}

	fieldValue.Set(cv)
	return nil
}

func (c *copyCreator_Struct) ensureValue() {
	if !c.v.IsValid() {
		if c.t.Kind() == reflect.Ptr {
			c.v = reflect.New(c.t.Elem())
		} else {
			c.v = reflect.New(c.t).Elem()
		}
		c.it = rprim.IndirectType(c.t)
	}
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
	c *Config
	t reflect.Type
	v reflect.Value
}

func (c *copyCreator_Map) Type() reflect.Type {
	return c.t
}

func (c *copyCreator_Map) SetCurrentValue(current reflect.Value) error {
	if current.Type() != c.t {
		return fmt.Errorf("Destination is not of the same type (%s -> %s)", current.Type().String(), c.t.String())
	}
	if !current.IsNil() {
		forceduplicate := !((c.c.Flags & XCF_OVERWRITE_EXISTING) == XCF_OVERWRITE_EXISTING)
		if !forceduplicate {
			c.v = current
		} else {
			// duplicate the map
			newValue, err := c.c.XCopyToNew(current, c.t)
			if err != nil {
				return err
			}
			c.v = newValue
		}
	}
	return nil
}

func (c *copyCreator_Map) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *copyCreator_Map) SetField(index reflect.Value, value reflect.Value) error {
	// convert index to the map index type
	mapindex, err := c.c.RprimConfig.Convert(index, rprim.IndirectType(c.t).Key())
	if err != nil {
		return err
	}

	c.ensureValue()

	// map items are not settable, if set, must be copied
	currentValue := reflect.Value{}
	if cur := c.v.MapIndex(mapindex); cur.IsValid() {
		if (c.c.Flags & XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE) == XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE {
			currentValue, err = c.c.XCopyToNew(cur, cur.Type())
			if err != nil {
				return err
			}
		} else {
			return errors.New("Map item is not settable, cannot set existing item")
		}
	}

	cv, err := c.c.XCopyUsingExistingIfValid(value, c.t.Elem(), currentValue)
	if err != nil {
		return err
	}

	c.v.SetMapIndex(mapindex, cv)
	return nil
}

func (c *copyCreator_Map) ensureValue() {
	if !c.v.IsValid() {
		c.v = reflect.MakeMap(c.t)
	}
}

func (c *copyCreator_Map) ensureValueOrZero() {
	if !c.v.IsValid() {
		//c.v = reflect.New(c.t).Elem()
		c.v = reflect.Zero(c.t)
	}
}

//
// Slice
//

type copyCreator_Slice struct {
	c *Config
	t reflect.Type
	v reflect.Value
}

func (c *copyCreator_Slice) Type() reflect.Type {
	return c.t
}

func (c *copyCreator_Slice) SetCurrentValue(current reflect.Value) error {
	if current.Type() != c.t {
		return fmt.Errorf("Destination is not of the same type (%s -> %s)", current.Type().String(), c.t.String())
	}
	if !current.IsNil() {
		forceduplicate := !((c.c.Flags & XCF_OVERWRITE_EXISTING) == XCF_OVERWRITE_EXISTING)
		if !forceduplicate {
			c.v = current
		} else {
			// duplicate the slice
			newValue, err := c.c.XCopyToNew(current, c.t)
			if err != nil {
				return err
			}
			c.v = newValue
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

	c.ensureValue()

	// Add zero values until the index
	for int(sliceindex.Int()) >= c.v.Len() {
		c.v = reflect.Append(c.v, reflect.Zero(c.t.Elem()))
	}

	currentValue := reflect.Value{}
	if int(sliceindex.Int()) < c.v.Len() {
		currentValue = c.v.Index(int(sliceindex.Int()))
	}

	cv, err := c.c.XCopyUsingExistingIfValid(value, c.t.Elem(), currentValue)
	if err != nil {
		return err
	}

	c.v.Index(int(sliceindex.Int())).Set(cv)
	return nil
}

func (c *copyCreator_Slice) ensureValue() {
	if !c.v.IsValid() {
		c.v = reflect.MakeSlice(c.t, 0, 0)
	}
}

func (c *copyCreator_Slice) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.Zero(c.t)
	}
}

//
// Primitive
//

type copyCreator_Primitive struct {
	c  *Config
	t  reflect.Type
	it reflect.Type
	v  reflect.Value
}

func (c *copyCreator_Primitive) Type() reflect.Type {
	return c.t
}

func (c *copyCreator_Primitive) SetCurrentValue(current reflect.Value) error {
	if current.Type() != c.t {
		return fmt.Errorf("Destination is not of the same type (%s -> %s)", current.Type().String(), c.t.String())
	}
	forceduplicate := !((c.c.Flags & XCF_OVERWRITE_EXISTING) == XCF_OVERWRITE_EXISTING)
	if !forceduplicate {
		c.v = current
		c.it = rprim.IndirectType(c.t)
	} else {
		if current.Kind() != reflect.Ptr || !current.IsNil() {
			// duplicate the primitive
			newValue, err := c.c.XCopyToNew(current, c.t)
			if err != nil {
				return err
			}
			c.v = newValue
			c.it = rprim.IndirectType(c.t)
		}
	}
	c.it = rprim.IndirectType(c.t)
	return nil
}

func (c *copyCreator_Primitive) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *copyCreator_Primitive) SetField(index reflect.Value, value reflect.Value) error {
	if index.IsValid() {
		return fmt.Errorf("Cannot set a primitive with an index")
	}

	c.ensureValue()
	var err error
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
			return errors.New("The primitive value is not settable, and the destination value is nil")
		} else if val.IsNil() {
			return errors.New("The primitive value is not settable, and the source value is nil")
		} else {
			c.v.Elem().Set(val.Elem())
		}
	} else {
		return errors.New("The primitive value is not settable")
	}
	return nil
}

func (c *copyCreator_Primitive) ensureValue() {
	if !c.v.IsValid() {
		c.v = reflect.New(c.t).Elem()
		c.it = rprim.IndirectType(c.t)
	}
}

func (c *copyCreator_Primitive) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.New(c.t).Elem()
	}
}
