package goxcopy

import (
	"errors"
	"reflect"
)

func XCopyToNew(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	stype := reflectTypeIndirect(src.Type())

	switch stype.Kind() {
	case reflect.Struct:
		return XCopyToNew_Struct(src, destType)
	case reflect.String, reflect.Int, reflect.Int16:
		return XCopyToNew_Primitive(src, destType)
	}

	return reflect.Value{}, errors.New("Unknown source type: " + stype.Kind().String())
}

func XCopyToNew_Struct(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	srcValue := reflect.Indirect(src)

	destCreator := XCopyGetCreator(destType)

	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		srcFieldType := srcValue.Type().Field(i)

		if srcFieldType.PkgPath != "" {
			// skip unexported fields
			continue
		}

		err := destCreator.SetField(reflect.ValueOf(srcFieldType.Name), srcField)
		if err != nil {
			return reflect.Value{}, err
		}
	}

	return destCreator.Create()
}

func XCopyToNew_Primitive(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	destCreator := XCopyGetCreator(destType)

	err := destCreator.SetField(reflect.Value{}, src)
	if err != nil {
		return reflect.Value{}, err
	}

	return destCreator.Create()
}

func reflectTypeIndirect(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}
