package goxcopy

import (
	"reflect"
	"testing"
)

type ST_Source struct {
	String1    string
	String2    *string
	Int1       int
	Int2       *int
	Float1     float32
	Float2     *float32
	Interface1 interface{}
}

func NewST_Source() *ST_Source {
	string2 := "2__string"
	int2 := 888
	float2 := float32(13.5)

	return &ST_Source{
		String1:    "1__string",
		String2:    &string2,
		Int1:       777,
		Int2:       &int2,
		Float1:     11.5,
		Float2:     &float2,
		Interface1: 99,
	}
}

type ST_Dest struct {
	String1    *string
	String2    string
	Int1       *int
	Int2       int
	Float1     *float32
	Float2     float32
	Interface1 string
	ExtraValue string
}

func TestStructToStruct(t *testing.T) {
	s := NewST_Source()

	gen, err := CopyToNew(s, reflect.TypeOf(&ST_Dest{}))
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.(*ST_Dest)
	if !issd {
		t.Fatal("Result is not *ST_Dest")
	}

	if s.String1 != *sd.String1 ||
		*s.String2 != sd.String2 ||
		s.Int1 != *sd.Int1 ||
		*s.Int2 != sd.Int2 ||
		s.Float1 != *sd.Float1 ||
		*s.Float2 != sd.Float2 ||
		sd.Interface1 != "99" {
		t.Fatal("Values are different")
	}
}

func TestStructToStructUsingExisting(t *testing.T) {
	s := NewST_Source()

	xsd := &ST_Dest{
		ExtraValue: "555444",
	}

	gen, err := CopyUsingExisting(s, xsd)
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.(*ST_Dest)
	if !issd {
		t.Fatal("Result is not *ST_Dest")
	}

	if s.String1 != *sd.String1 ||
		*s.String2 != sd.String2 ||
		s.Int1 != *sd.Int1 ||
		*s.Int2 != sd.Int2 ||
		s.Float1 != *sd.Float1 ||
		*s.Float2 != sd.Float2 ||
		sd.Interface1 != "99" {
		t.Fatal("Values are different")
	}

	if sd.ExtraValue != "555444" {
		t.Fatal("Existing value was not kept")
	}

	if xsd.String2 != "" {
		t.Fatal("The original struct should not have been changed")
	}
}

func TestStructToStructExistingInplace(t *testing.T) {
	s := NewST_Source()

	xsd := &ST_Dest{
		ExtraValue: "555444",
	}

	err := CopyToExisting(s, xsd)
	if err != nil {
		t.Fatal(err)
	}

	if xsd.String2 != *s.String2 {
		t.Fatal("The original struct should been changed")
	}
}

func TestStructToMap(t *testing.T) {
	s := NewST_Source()

	gen, err := CopyToNew(s, reflect.TypeOf(map[string]interface{}{}))
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.(map[string]interface{})
	if !issd {
		t.Fatal("Result is not map[string]interface{}")
	}

	if s.String1 != sd["String1"] ||
		s.String2 != sd["String2"] ||
		s.Int1 != sd["Int1"] ||
		s.Int2 != sd["Int2"] ||
		s.Float1 != sd["Float1"] ||
		s.Float2 != sd["Float2"] ||
		s.Interface1 != sd["Interface1"] {
		t.Fatal("Values are different")
	}
}

func TestStructToPrimitive(t *testing.T) {
	s := NewST_Source()

	_, err := CopyToNew(s, reflect.TypeOf(0))
	if err == nil {
		t.Fatal("Should not allow to copy struct to primitive")
	}
}

func TestStructTag(t *testing.T) {
	type s_type struct {
		Value1 string `goxcopy:"New_value1"`
		Value2 string
	}

	type d_type struct {
		Value1     string
		New_value1 string
		New_value2 string `goxcopy:"Value2"`
	}

	s := &s_type{
		Value1: "u_value_1",
		Value2: "u_value_2",
	}

	d, err := CopyToNew(s, reflect.TypeOf(&d_type{}))
	if err != nil {
		t.Fatal(err)
	}

	dt, ok := d.(*d_type)
	if !ok {
		t.Fatal("Result isn't *d_type")
	}

	if dt.Value1 != "" || dt.New_value1 != "u_value_1" || dt.New_value2 != "u_value_2" {
		t.Fatal("Value not set as expected")
	}
}

func TestStructOverwriteNilPointer(t *testing.T) {
	type sx struct {
		Value1 string
		Value2 *sx
	}

	src := &sx{
		Value1: "x_value1",
	}
	dst := &sx{}

	err := CopyToExisting(src, dst)
	if err != nil {
		t.Fatal(err)
	}

	if src.Value1 != dst.Value1 || dst.Value2 != nil {
		t.Fatal("Structs have different values")
	}
}

func TestStructOverwritePointerIndirection(t *testing.T) {
	type s struct {
		SValue1 string
	}

	type sx struct {
		Value1 string
		Value2 *s
	}

	type sy struct {
		Value1 string
		Value2 **s
	}

	src := &sx{
		Value1: "x_value1",
		Value2: &s{
			SValue1: "s_value",
		},
	}
	dst := &sy{}

	err := CopyToExisting(src, dst)
	if err != nil {
		t.Fatal(err)
	}

	if src.Value1 != dst.Value1 || dst.Value2 == nil || src.Value2.SValue1 != (*dst.Value2).SValue1 {
		t.Fatal("Structs have different values")
	}
}
