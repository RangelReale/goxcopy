package goxcopy

import (
	"fmt"
	"github.com/RangelReale/rprim"
	"reflect"
)

func XCopyToNew(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	stype := rprim.IndirectType(src.Type())

	switch stype.Kind() {
	case reflect.Struct:
		return xCopyToNew_Struct(src, destType)
	case reflect.Map:
		return xCopyToNew_Map(src, destType)
	case reflect.Slice:
		return xCopyToNew_Slice(src, destType)
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String, reflect.Interface:
		return xCopyToNew_Primitive(src, destType)
	}
	return reflect.Value{}, fmt.Errorf("Kind not supported: %s", stype.Kind().String())
}

//
// Struct
//
func xCopyToNew_Struct(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	srcValue := reflect.Indirect(src)

	destCreator, err := XCopyGetCreator(destType)
	if err != nil {
		return reflect.Value{}, err
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
func xCopyToNew_Map(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	srcValue := reflect.Indirect(src)

	destCreator, err := XCopyGetCreator(destType)
	if err != nil {
		return reflect.Value{}, err
	}

	for _, k := range srcValue.MapKeys() {
		srcField := srcValue.MapIndex(k)

		err := destCreator.SetField(k, srcField)
		if err != nil {
			kstr, err := rprim.ConvertToString(k)
			if err != nil {
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
func xCopyToNew_Slice(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	srcValue := reflect.Indirect(src)

	destCreator, err := XCopyGetCreator(destType)
	if err != nil {
		return reflect.Value{}, err
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
func xCopyToNew_Primitive(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	destCreator, err := XCopyGetCreator(destType)
	if err != nil {
		return reflect.Value{}, err
	}

	err = destCreator.SetField(reflect.Value{}, src)
	if err != nil {
		return reflect.Value{}, err
	}

	return destCreator.Create()
}
