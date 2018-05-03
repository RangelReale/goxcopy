package goxcopy

import (
	"fmt"
	"reflect"

	"github.com/RangelReale/rprim"
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

const (
	// Whether to overwrite existing items instead of returning a copy
	XCF_OVERWRITE_EXISTING = 1
	// Whether to allow creating a new item from a copy if the item is not assignable (only if using existing)
	XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE = 2
	// Return error if source struct founds no corresponding field on the target
	XCF_ERROR_IF_STRUCT_FIELD_MISSING = 4
	// Disable special handling of map[XXX]interface{} target, which can create a new inner map of the same type
	// if the source have fields.
	XCF_DISABLE_MAPOFINTERFACE_TARGET_RECURSION = 8
)

//
// Copy configuration
//
type Config struct {
	// XCF_* flag bitmask
	Flags uint
	// Name of the struct field tag to find out the element name (default: goxcopy)
	StructTagName string
	// Field map
	FieldMap map[string]*FieldMap
	// Configuration of the primitive type converter
	RprimConfig *rprim.Config
}

// Creates a new default Config
func NewConfig() *Config {
	return &Config{
		StructTagName: "goxcopy",
		RprimConfig:   rprim.NewConfig(),
	}
}

// Duplicates the Config
func (c *Config) Dup() *Config {
	ret := &Config{
		Flags:         c.Flags,
		StructTagName: c.StructTagName,
		RprimConfig:   c.RprimConfig.Dup(),
	}
	if c.FieldMap != nil {
		ret.FieldMap = make(map[string]*FieldMap)
		for fn, fv := range c.FieldMap {
			ret.FieldMap[fn] = fv
		}
	}
	return ret
}

// Reset the config flags
func (c *Config) SetFlags(flags uint) *Config {
	c.Flags = flags
	return c
}

// Add flag bitmask to the config
func (c *Config) AddFlags(flags uint) *Config {
	c.Flags |= flags
	return c
}

// Set the rprim config
func (c *Config) SetRprimConfig(rc *rprim.Config) *Config {
	c.RprimConfig = rc
	return c
}

func (c *Config) SetFieldMap(fm map[string]*FieldMap) *Config {
	c.FieldMap = fm
	return c
}

func (c *Config) GetFieldMap(fieldname string) *FieldMap {
	if c.FieldMap == nil {
		return nil
	}
	if fm, ok := c.FieldMap[fieldname]; ok {
		return fm
	}
	return nil
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
	_, err := c.Dup().AddFlags(XCF_OVERWRITE_EXISTING).XCopyUsingExisting(NewContext(), reflect.ValueOf(src), reflect.ValueOf(currentValue))
	return err
}

// Copy a source variable to a new instance of the passed type.
// The src variable is never changed in any circunstance.
func (c *Config) XCopyToNew(ctx *Context, src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	return c.XCopyUsingExistingIfValid(ctx, src, destType, reflect.Value{})
}

// Copy a source variable to a new instance of the type of the passed value.
// The passed variable is used to initialize the new instance value, but is not
// changed in any way.
// The src and currentValue variable are never changed in any circunstance.
func (c *Config) XCopyUsingExisting(ctx *Context, src reflect.Value, currentValue reflect.Value) (reflect.Value, error) {
	return c.XCopyUsingExistingIfValid(ctx, src, reflect.TypeOf(currentValue.Interface()), currentValue)
}

// Copy a source variable to a destination variable, overwriting it.
// The destination variable must be settable.
// This is an alias for "CopyToExisting"
// The src variable is never changed in any circunstance.
func (c *Config) XCopyToExisting(ctx *Context, src reflect.Value, currentValue reflect.Value) error {
	_, err := c.Dup().AddFlags(XCF_OVERWRITE_EXISTING).XCopyUsingExistingIfValid(ctx, src, reflect.TypeOf(currentValue.Interface()), currentValue)
	return err
}

// The underling function that does the other functions work.
func (c *Config) XCopyUsingExistingIfValid(ctx *Context, src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	skind := rprim.UnderliningValueKind(src)

	switch skind {
	case reflect.Struct:
		return c.copyTo_Struct(ctx, src, destType, currentValue)
	case reflect.Map:
		return c.copyTo_Map(ctx, src, destType, currentValue)
	case reflect.Slice:
		return c.copyTo_Slice(ctx, src, destType, currentValue)
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String, reflect.Interface:
		return c.copyTo_Primitive(ctx, src, destType, currentValue)
	}
	return reflect.Value{}, newError(fmt.Errorf("Kind not supported: %s", skind.String()), ctx.Dup())
}

//
// Struct copy source
//
func (c *Config) copyTo_Struct(ctx *Context, src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	srcValue := rprim.UnderliningValue(src)

	destCreator, err := c.XCopyGetCreator(ctx, destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	if srcValue.Kind() != reflect.Ptr || !srcValue.IsNil() {
		for i := 0; i < srcValue.NumField(); i++ {
			srcField := srcValue.Field(i)
			srcFieldType := srcValue.Type().Field(i)

			if srcFieldType.PkgPath != "" {
				// skip unexported fields
				continue
			}

			targetFieldName := srcFieldType.Name

			// check for the struct tag and change the field name if requested
			tag_fields := c.GetStructTagFields(srcFieldType)
			if len(tag_fields) > 0 {
				if tag_fields[0] == "-" {
					targetFieldName = "" // skip
				} else {
					targetFieldName = tag_fields[0]
				}
			}

			if targetFieldName != "" {
				// check the field map for this field
				if fieldmap := c.GetFieldMap(ctx.FieldsAsStringAppending(reflect.ValueOf(targetFieldName))); fieldmap != nil {
					if fieldmap.Fieldname != nil {
						targetFieldName = *fieldmap.Fieldname
					}
				}
			}

			if targetFieldName != "" {
				// set the field on the creator
				fv := reflect.ValueOf(targetFieldName)

				ctx.PushField(fv)
				err := destCreator.SetField(fv, srcField)
				ctx.PopField()

				if err != nil {
					return reflect.Value{}, err
				}
			}
		}
	}

	return destCreator.Create()
}

//
// Map copy source
//
func (c *Config) copyTo_Map(ctx *Context, src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	srcValue := rprim.UnderliningValue(src)

	destCreator, err := c.XCopyGetCreator(ctx, destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	if srcValue.Kind() != reflect.Ptr || !srcValue.IsNil() {
		for _, k := range srcValue.MapKeys() {
			srcField := srcValue.MapIndex(k)

			kindex := k
			// check the field map for this field
			if fieldmap := c.GetFieldMap(ctx.FieldsAsStringAppending(kindex)); fieldmap != nil {
				if fieldmap.Fieldname != nil {
					kindex = reflect.ValueOf(*fieldmap.Fieldname)
				}
			}

			// set the value on the creator
			ctx.PushField(kindex)
			err := destCreator.SetField(kindex, srcField)
			ctx.PopField()

			if err != nil {
				return reflect.Value{}, err
			}
		}
	}

	return destCreator.Create()
}

//
// Slice copy source
//
func (c *Config) copyTo_Slice(ctx *Context, src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	srcValue := rprim.UnderliningValue(src)

	destCreator, err := c.XCopyGetCreator(ctx, destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	if srcValue.Kind() != reflect.Ptr || !srcValue.IsNil() {
		for i := 0; i < srcValue.Len(); i++ {
			srcField := srcValue.Index(i)

			fv := reflect.ValueOf(i)

			fvindex := fv
			// check the field map for this field
			if fieldmap := c.GetFieldMap(ctx.FieldsAsStringAppending(fv)); fieldmap != nil {
				if fieldmap.Fieldname != nil {
					fvindex = reflect.ValueOf(*fieldmap.Fieldname)
				}
			}

			// set the value on the creator
			ctx.PushField(fvindex)
			err := destCreator.SetField(fvindex, srcField)
			ctx.PopField()

			if err != nil {
				return reflect.Value{}, err
			}
		}
	}

	return destCreator.Create()
}

//
// Primitive copy source
//
func (c *Config) copyTo_Primitive(ctx *Context, src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	destCreator, err := c.XCopyGetCreator(ctx, destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	// set the value on the creator. As primitives don't have fields, the index is passed as an INVALID reflect.Value.
	err = destCreator.SetField(reflect.Value{}, src)
	if err != nil {
		return reflect.Value{}, err
	}

	return destCreator.Create()
}
