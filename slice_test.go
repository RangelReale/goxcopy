package goxcopy

import (
	"reflect"
	"testing"
)

func NewLT_Source() []interface{} {
	return []interface{}{
		"1__string",
		777,
		float32(11.5),
	}
}

func TestSliceToSlice(t *testing.T) {
	s := NewLT_Source()

	gen, err := CopyToNew(s, reflect.TypeOf(s))
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.([]interface{})
	if !issd {
		t.Fatal("Result is not []interface{}")
	}

	if s[0] != sd[0] ||
		s[1] != sd[1] ||
		s[2] != sd[2] {
		t.Fatal("Values are different")
	}
}

func TestSliceToSliceExisting(t *testing.T) {
	s := NewLT_Source()

	xsd := []interface{}{
		nil,
		nil,
		nil,
		"Existing item",
	}

	gen, err := CopyUsingExisting(s, xsd)
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.([]interface{})
	if !issd {
		t.Fatal("Result is not []interface{}")
	}

	if s[0] != sd[0] ||
		s[1] != sd[1] ||
		s[2] != sd[2] ||
		sd[3].(string) != "Existing item" {
		t.Fatal("Values are different")
	}
}

func TestSliceToSliceExistingInplace(t *testing.T) {
	s := NewLT_Source()

	xsd := []interface{}{
		nil,
		nil,
		nil,
		"Existing item",
	}

	err := NewConfig().CopyToExisting(s, &xsd)
	if err != nil {
		t.Fatal(err)
	}

	if s[0] != xsd[0] ||
		s[1] != xsd[1] ||
		s[2] != xsd[2] ||
		xsd[3].(string) != "Existing item" {
		t.Fatal("Values are different")
	}
}

func TestSliceToMap(t *testing.T) {
	s := NewLT_Source()

	gen, err := CopyToNew(s, reflect.TypeOf(map[string]interface{}{}))
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.(map[string]interface{})
	if !issd {
		t.Fatal("Result is not map[string]interface{}")
	}

	if s[0] != sd["0"] ||
		s[1] != sd["1"] ||
		s[2] != sd["2"] {
		t.Fatal("Values are different")
	}
}

func TestSliceToMapStringValue(t *testing.T) {
	s := NewLT_Source()

	gen, err := CopyToNew(s, reflect.TypeOf(map[string]string{}))
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.(map[string]string)
	if !issd {
		t.Fatal("Result is not map[string]interface{}")
	}

	if "1__string" != sd["0"] ||
		"777" != sd["1"] {
		t.Fatal("Values are different")
	}
}

func TestSliceUnderlining(t *testing.T) {
	src := []string{
		"value1",
		"value2",
	}
	var dst ****[]string

	err := CopyToExisting(src, &dst)
	if err != nil {
		t.Fatal(err)
	}

	if dst == nil || *dst == nil || **dst == nil || ***dst == nil || ****dst == nil {
		t.Fatal("Values are different")
	}

	xdst := ****dst
	if len(xdst) != 2 || xdst[0] != "value1" {
		t.Fatal("Values are different")
	}
}
