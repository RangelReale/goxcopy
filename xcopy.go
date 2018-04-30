package goxcopy

import (
	"fmt"
	"reflect"

	"github.com/RangelReale/rprim"
)

func CopyToNew(src interface{}, destType reflect.Type) (interface{}, error) {
	return NewConfig().CopyToNew(src, destType)
}

func CopyFromExisting(src interface{}, currentValue interface{}) (interface{}, error) {
	return NewConfig().CopyFromExisting(src, currentValue)
}

func CopyToExisting(src interface{}, currentValue interface{}) error {
	return NewConfig().CopyToExisting(src, currentValue)
}

func XCopyToNew(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	return NewConfig().XCopyToNew(src, destType)
}

func XCopyFromExisting(src reflect.Value, currentValue reflect.Value) (reflect.Value, error) {
	return NewConfig().XCopyFromExisting(src, currentValue)
}

func XCopyToExisting(src reflect.Value, currentValue reflect.Value) error {
	return NewConfig().XCopyToExisting(src, currentValue)
}

const (
	// Whether to overwrite existing items instead of returning a copy
	XCF_OVERWRITE_EXISTING = 1
	// Whether to allow creating a new item from a copy if the item is not assignable (only if using existing)
	XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE = 2
	//
	XCF_ERROR_IF_STRUCT_FIELD_MISSING = 4
)

//
// Config
//
type Config struct {
	Flags       uint
	RprimConfig *rprim.Config
}

func NewConfig() *Config {
	return &Config{
		RprimConfig: rprim.NewConfig(),
	}
}

func (c Config) Dup() *Config {
	return &Config{
		Flags:       c.Flags,
		RprimConfig: c.RprimConfig.Dup(),
	}
}

func (c *Config) SetFlags(flags uint) *Config {
	c.Flags = flags
	return c
}

func (c *Config) AddFlags(flags uint) *Config {
	c.Flags |= flags
	return c
}

func (c *Config) SetRprimConfig(rc *rprim.Config) *Config {
	c.RprimConfig = rc
	return c
}

func (c *Config) CopyToNew(src interface{}, destType reflect.Type) (interface{}, error) {
	ret, err := c.XCopyToNew(reflect.ValueOf(src), destType)
	if err != nil {
		return nil, err
	}
	return ret.Interface(), nil
}

func (c *Config) CopyFromExisting(src interface{}, currentValue interface{}) (interface{}, error) {
	ret, err := c.XCopyFromExisting(reflect.ValueOf(src), reflect.ValueOf(currentValue))
	if err != nil {
		return nil, err
	}
	return ret.Interface(), nil
}

func (c *Config) CopyToExisting(src interface{}, currentValue interface{}) error {
	_, err := c.Dup().AddFlags(XCF_OVERWRITE_EXISTING).XCopyFromExisting(reflect.ValueOf(src), reflect.ValueOf(currentValue))
	return err
}

func (c *Config) XCopyToNew(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	return c.XCopyFromExistingIfValid(src, destType, reflect.Value{})
}

func (c *Config) XCopyFromExisting(src reflect.Value, currentValue reflect.Value) (reflect.Value, error) {
	return c.XCopyFromExistingIfValid(src, reflect.TypeOf(currentValue.Interface()), currentValue)
}

func (c *Config) XCopyToExisting(src reflect.Value, currentValue reflect.Value) error {
	_, err := c.Dup().AddFlags(XCF_OVERWRITE_EXISTING).XCopyFromExistingIfValid(src, reflect.TypeOf(currentValue.Interface()), currentValue)
	return err
}

func (c *Config) XCopyFromExistingIfValid(src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	stype := rprim.IndirectType(src.Type())

	switch stype.Kind() {
	case reflect.Struct:
		return c.copyToNew_Struct(src, destType, currentValue)
	case reflect.Map:
		return c.copyToNew_Map(src, destType, currentValue)
	case reflect.Slice:
		return c.copyToNew_Slice(src, destType, currentValue)
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String, reflect.Interface:
		return c.copyToNew_Primitive(src, destType, currentValue)
	}
	return reflect.Value{}, fmt.Errorf("Kind not supported: %s", stype.Kind().String())
}

//
// Struct
//
func (c *Config) copyToNew_Struct(src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	srcValue := reflect.Indirect(src)

	destCreator, err := c.XCopyGetCreator(destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		srcFieldType := srcValue.Type().Field(i)

		if srcFieldType.PkgPath != "" {
			// skip unexported fields
			continue
		}

		err := destCreator.SetField(reflect.ValueOf(srcFieldType.Name), srcField)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("Error copying from struct field %s: %v", srcFieldType.Name, err)
		}
	}

	return destCreator.Create()
}

//
// Map
//
func (c *Config) copyToNew_Map(src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	srcValue := reflect.Indirect(src)

	destCreator, err := c.XCopyGetCreator(destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	for _, k := range srcValue.MapKeys() {
		srcField := srcValue.MapIndex(k)

		err := destCreator.SetField(k, srcField)
		if err != nil {
			kstr, kerr := c.RprimConfig.ConvertToString(k)
			if kerr != nil {
				kstr = "<unknown>"
			}
			return reflect.Value{}, fmt.Errorf("Error copying from map index %s: %v", kstr, err)
		}
	}

	return destCreator.Create()
}

//
// Slice
//
func (c *Config) copyToNew_Slice(src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	srcValue := reflect.Indirect(src)

	destCreator, err := c.XCopyGetCreator(destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	for i := 0; i < srcValue.Len(); i++ {
		srcField := srcValue.Index(i)

		err := destCreator.SetField(reflect.ValueOf(i), srcField)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("Error copying from slice index %d: %v", i, err)
		}
	}

	return destCreator.Create()
}

//
// Primitive
//
func (c *Config) copyToNew_Primitive(src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	destCreator, err := c.XCopyGetCreator(destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	err = destCreator.SetField(reflect.Value{}, src)
	if err != nil {
		return reflect.Value{}, err
	}

	return destCreator.Create()
}
