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

func XCopyGetCreator(t reflect.Type) XCopyCreator {
	switch reflectTypeIndirect(t).Kind() {
	case reflect.Struct:
		return &XCopyCreator_Struct{t: t}
	case reflect.Map:
		return &XCopyCreator_Map{t: t}
	case reflect.String, reflect.Int, reflect.Int32, reflect.Float32, reflect.Interface:
		return &XCopyCreator_Primitive{t: t}
	}
	panic("Must handle all reflect kinds: " + t.Kind().String())
}

//
// Struct
//

type XCopyCreator_Struct struct {
	t  reflect.Type
	it reflect.Type
	v  reflect.Value
}

func (c *XCopyCreator_Struct) Type() reflect.Type {
	return c.t
}

func (c *XCopyCreator_Struct) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *XCopyCreator_Struct) SetField(index reflect.Value, value reflect.Value) error {
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

func (c *XCopyCreator_Struct) ensureValue() {
	if !c.v.IsValid() {
		if c.t.Kind() == reflect.Ptr {
			c.v = reflect.New(c.t.Elem())
		} else {
			c.v = reflect.New(c.t).Elem()
		}
		c.it = reflectTypeIndirect(c.t)
	}
}

func (c *XCopyCreator_Struct) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.New(c.t).Elem()
	}
}

//
// Map
//

type XCopyCreator_Map struct {
	t reflect.Type
	v reflect.Value
}

func (c *XCopyCreator_Map) Type() reflect.Type {
	return c.t
}

func (c *XCopyCreator_Map) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *XCopyCreator_Map) SetField(index reflect.Value, value reflect.Value) error {
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

func (c *XCopyCreator_Map) ensureValue() {
	if !c.v.IsValid() {
		c.v = reflect.MakeMap(c.t)
	}
}

func (c *XCopyCreator_Map) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.New(c.t).Elem()
	}
}

//
// Primitive
//

type XCopyCreator_Primitive struct {
	t  reflect.Type
	it reflect.Type
	v  reflect.Value
}

func (c *XCopyCreator_Primitive) Type() reflect.Type {
	return c.t
}

func (c *XCopyCreator_Primitive) Create() (reflect.Value, error) {
	c.ensureValueOrZero()
	return c.v, nil
}

func (c *XCopyCreator_Primitive) SetField(index reflect.Value, value reflect.Value) error {
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

func (c *XCopyCreator_Primitive) ensureValue() {
	if !c.v.IsValid() {
		if c.t.Kind() == reflect.Ptr {
			c.v = reflect.New(c.t.Elem())
		} else {
			c.v = reflect.New(c.t).Elem()
		}
		c.it = reflectTypeIndirect(c.t)
	}
}

func (c *XCopyCreator_Primitive) ensureValueOrZero() {
	if !c.v.IsValid() {
		c.v = reflect.New(c.t).Elem()
	}
}
