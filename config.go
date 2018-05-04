package goxcopy

import (
	"fmt"
	"reflect"

	"github.com/RangelReale/rprim"
)

const (
	// Whether to overwrite existing items instead of returning a copy
	XCF_OVERWRITE_EXISTING = 1
	// Whether to allow creating a new non-primitive item from a copy if the item is not assignable (only if using existing)
	XCF_ALLOW_DUPLICATING_IF_NOT_SETTABLE = 2
	// Whether to deny creating a new primitive item from a copy if the item is not assignable (only if using existing)
	XCF_DENY_DUPLICATING_PRIMITIVE_IF_NOT_SETTABLE = 4
	// Return error if source struct founds no corresponding field on the target
	XCF_ERROR_IF_STRUCT_FIELD_MISSING = 8
	// Disable special handling of map[XXX]interface{} target, which can create a new inner map of the same type
	// if the source have fields.
	XCF_DISABLE_MAPOFINTERFACE_TARGET_RECURSION = 16
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
	Callback    Callback
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
		Callback:      c.Callback,
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

// Set the callback
func (c *Config) SetCallback(callback Callback) *Config {
	c.Callback = callback
	return c
}

// The underling function that does the other functions work.
func (c *Config) internalXCopyUsingExistingIfValid(ctx *Context, src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	skind := rprim.UnderliningValueKind(src)

	switch skind {
	case reflect.Struct:
		return c.copyTo_Struct(ctx, src, destType, currentValue)
	case reflect.Map:
		return c.copyTo_Map(ctx, src, destType, currentValue)
	case reflect.Slice, reflect.Array:
		return c.copyTo_Slice(ctx, src, destType, currentValue)
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String, reflect.Interface:
		return c.copyTo_Primitive(ctx, src, destType, currentValue)
	}
	return reflect.Value{}, newError(fmt.Errorf("Kind not supported: %s", skind.String()), ctx)
}

//
// Struct copy source
//
func (c *Config) copyTo_Struct(ctx *Context, src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	srcValue := rprim.UnderliningValue(src)

	destCreator, err := c.GetCreator(ctx, destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	if !destCreator.TryFastCopy(src) {
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
					c.callbackPushField(ctx, fv, src, destCreator) // callback

					err := destCreator.SetField(fv, srcField)

					ctx.PopField()
					c.callbackPopField(ctx, fv, src, destCreator) // callback

					if err != nil {
						return reflect.Value{}, err
					}
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

	destCreator, err := c.GetCreator(ctx, destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	if !destCreator.TryFastCopy(src) {
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
				c.callbackPushField(ctx, kindex, src, destCreator) // callback

				err := destCreator.SetField(kindex, srcField)

				ctx.PopField()
				c.callbackPopField(ctx, kindex, src, destCreator) // callback

				if err != nil {
					return reflect.Value{}, err
				}
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

	destCreator, err := c.GetCreator(ctx, destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	if !destCreator.TryFastCopy(src) {
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
				c.callbackPushField(ctx, fvindex, src, destCreator) // callback

				err := destCreator.SetField(fvindex, srcField)

				ctx.PopField()
				c.callbackPopField(ctx, fvindex, src, destCreator) // callback

				if err != nil {
					return reflect.Value{}, err
				}
			}
		}
	}

	return destCreator.Create()
}

//
// Primitive copy source
//
func (c *Config) copyTo_Primitive(ctx *Context, src reflect.Value, destType reflect.Type, currentValue reflect.Value) (reflect.Value, error) {
	destCreator, err := c.GetCreator(ctx, destType)
	if err != nil {
		return reflect.Value{}, err
	}
	if currentValue.IsValid() {
		if err := destCreator.SetCurrentValue(currentValue); err != nil {
			return reflect.Value{}, err
		}
	}

	c.callbackBeforeSetValue(ctx, src, destCreator, currentValue) // callback

	if !destCreator.TryFastCopy(src) {
		// set the value on the creator. As primitives don't have fields, the index is passed as an INVALID reflect.Value.
		err = destCreator.SetField(reflect.Value{}, src)
		if err != nil {
			return reflect.Value{}, err
		}
	}

	c.callbackAfterSetValue(ctx, src, destCreator, currentValue) // callback

	return destCreator.Create()
}

// Callback helpers

func (c *Config) callbackBeginNew(ctx *Context, src reflect.Value, destType reflect.Type) {
	if c.Callback != nil {
		c.Callback.BeginNew(ctx, src, destType)
	}
}

func (c *Config) callbackEndNew(ctx *Context, src reflect.Value, destType reflect.Type) {
	if c.Callback != nil {
		c.Callback.EndNew(ctx, src, destType)
	}
}

func (c *Config) callbackPushField(ctx *Context, fieldname reflect.Value, src reflect.Value, dest Creator) {
	if c.Callback != nil {
		c.Callback.PushField(ctx, fieldname, src, dest)
	}
}

func (c *Config) callbackPopField(ctx *Context, fieldname reflect.Value, src reflect.Value, dest Creator) {
	if c.Callback != nil {
		c.Callback.PopField(ctx, fieldname, src, dest)
	}
}

func (c *Config) callbackBeforeSetValue(ctx *Context, src reflect.Value, dest Creator, value reflect.Value) {
	if c.Callback != nil {
		c.Callback.BeforeSetValue(ctx, src, dest, value)
	}
}

func (c *Config) callbackAfterSetValue(ctx *Context, src reflect.Value, dest Creator, value reflect.Value) {
	if c.Callback != nil {
		c.Callback.AfterSetValue(ctx, src, dest, value)
	}
}
