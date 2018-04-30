package goxcopy

import (
	"fmt"
	"reflect"

	"github.com/RangelReale/rprim"
)

type XCopyCreator interface {
	Type() reflect.Type
	Create() (reflect.Value, error)
	SetField(index reflect.Value, value reflect.Value) error
}

func XCopyGetCreator(t reflect.Type) (XCopyCreator, error) {
	switch rprim.IndirectType(t).Kind() {
	case reflect.Struct:
		return &xCopyCreator_Struct{t: t}, nil
	case reflect.Map:
		return &xCopyCreator_Map{t: t}, nil
	case reflect.Slice:
		return &xCopyCreator_Slice{t: t}, nil
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String, reflect.Interface:
		return &xCopyCreator_Primitive{t: t}, nil
	}
	return nil, fmt.Errorf("Kind not supported: %s", t.Kind().String())
}

//
// Struct
//

type xCopyCreator_Struct struct {
	t  reflect.Type
	it reflect.Type
	v  reflect.Value
}

func (c *xCopyCreator_Struct) Type() reflect.Type {
	return c.t
}

func (c *xCopyCreator_Struct) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *xCopyCreator_Struct) SetField(index reflect.Value, value reflect.Value) error {
	fieldname, err := rprim.ConvertToString(index)
	if err != nil {
		return err
	}

	c.ensureValue()

	fieldType, ok := c.it.FieldByName(fieldname)
	if !ok {
		return nil // TODO: handle possible error
	}

	fieldValue := reflect.Indirect(c.v).FieldByName(fieldname)

	cv, err := XCopyToNew(value, fieldType.Type)
	if err != nil {
		return err
	}

	fieldValue.Set(cv)
	return nil
}

func (c *xCopyCreator_Struct) ensureValue() {
	if !c.v.IsValid() {
		if c.t.Kind() == reflect.Ptr {
			c.v = reflect.New(c.t.Elem())
		} else {
			c.v = reflect.New(c.t).Elem()
		}
		c.it = rprim.IndirectType(c.t)
	}
}

func (c *xCopyCreator_Struct) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.New(c.t).Elem()
	}
}

//
// Map
//

type xCopyCreator_Map struct {
	t reflect.Type
	v reflect.Value
}

func (c *xCopyCreator_Map) Type() reflect.Type {
	return c.t
}

func (c *xCopyCreator_Map) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *xCopyCreator_Map) SetField(index reflect.Value, value reflect.Value) error {
	// convert index to the map index type
	mapindex, err := rprim.Convert(index, rprim.IndirectType(c.t).Key())
	if err != nil {
		return err
	}

	c.ensureValue()

	cv, err := XCopyToNew(value, c.t.Elem())
	if err != nil {
		return err
	}

	c.v.SetMapIndex(mapindex, cv)
	return nil
}

func (c *xCopyCreator_Map) ensureValue() {
	if !c.v.IsValid() {
		c.v = reflect.MakeMap(c.t)
	}
}

func (c *xCopyCreator_Map) ensureValueOrZero() {
	if !c.v.IsValid() {
		//c.v = reflect.New(c.t).Elem()
		c.v = reflect.Zero(c.t)
	}
}

//
// Slice
//

type xCopyCreator_Slice struct {
	t reflect.Type
	v reflect.Value
}

func (c *xCopyCreator_Slice) Type() reflect.Type {
	return c.t
}

func (c *xCopyCreator_Slice) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *xCopyCreator_Slice) SetField(index reflect.Value, value reflect.Value) error {
	// convert index to int
	sliceindex, err := rprim.Convert(index, reflect.TypeOf(0))
	if err != nil {
		return err
	}

	c.ensureValue()

	// Add zero values until the index
	for int(sliceindex.Int()) >= c.v.Len() {
		c.v = reflect.Append(c.v, reflect.Zero(c.t.Elem()))
	}

	cv, err := XCopyToNew(value, c.t.Elem())
	if err != nil {
		return err
	}

	c.v.Index(int(sliceindex.Int())).Set(cv)
	return nil
}

func (c *xCopyCreator_Slice) ensureValue() {
	if !c.v.IsValid() {
		c.v = reflect.MakeSlice(c.t, 0, 0)
	}
}

func (c *xCopyCreator_Slice) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.Zero(c.t)
	}
}

//
// Primitive
//

type xCopyCreator_Primitive struct {
	t  reflect.Type
	it reflect.Type
	v  reflect.Value
}

func (c *xCopyCreator_Primitive) Type() reflect.Type {
	return c.t
}

func (c *xCopyCreator_Primitive) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *xCopyCreator_Primitive) SetField(index reflect.Value, value reflect.Value) error {
	if index.IsValid() {
		return fmt.Errorf("Cannot set a primitive with an index")
	}

	c.ensureValue()
	var err error
	c.v, err = rprim.Convert(value, c.t)
	if err != nil {
		return err
	}
	return nil
}

func (c *xCopyCreator_Primitive) ensureValue() {
	if !c.v.IsValid() {
		if c.t.Kind() == reflect.Ptr {
			c.v = reflect.New(c.t.Elem())
		} else {
			c.v = reflect.New(c.t).Elem()
		}
		c.it = rprim.IndirectType(c.t)
	}
}

func (c *xCopyCreator_Primitive) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.New(c.t).Elem()
	}
}
