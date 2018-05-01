package goxcopy

import (
	"reflect"
	"testing"
)

func NewMT_Source() map[string]interface{} {
	return map[string]interface{}{
		"String1": "1__string",
		"Int1":    777,
		"Float1":  float32(11.5),
	}
}

func TestMapToMap(t *testing.T) {
	s := NewMT_Source()

	gen, err := CopyToNew(s, reflect.TypeOf(s))
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.(map[string]interface{})
	if !issd {
		t.Fatal("Result is not map[string]interface{}")
	}

	if s["String1"] != sd["String1"] ||
		s["Int1"] != sd["Int1"] ||
		s["Float1"] != sd["Float1"] {
		t.Fatal("Values are different")
	}
}

func TestMapToMapExisting(t *testing.T) {
	s := NewMT_Source()

	xsd := map[string]interface{}{
		"oldString1": "old___string1",
	}

	gen, err := CopyUsingExisting(s, xsd)
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.(map[string]interface{})
	if !issd {
		t.Fatal("Result is not map[string]interface{}")
	}

	if s["String1"] != sd["String1"] ||
		s["Int1"] != sd["Int1"] ||
		s["Float1"] != sd["Float1"] ||
		sd["oldString1"].(string) != "old___string1" {
		t.Fatal("Values are different")
	}

	if _, isset := xsd["String1"]; isset {
		t.Fatal("Source map should not been changed")
	}
}

func TestMapToMapExistingInplace(t *testing.T) {
	s := NewMT_Source()

	xsd := map[string]interface{}{
		"oldString1": "old___string1",
	}

	_, err := NewConfig().SetFlags(XCF_OVERWRITE_EXISTING).
		CopyUsingExisting(s, xsd)
	if err != nil {
		t.Fatal(err)
	}

	if s["String1"] != xsd["String1"] ||
		s["Int1"] != xsd["Int1"] ||
		s["Float1"] != xsd["Float1"] ||
		xsd["oldString1"].(string) != "old___string1" {
		t.Fatal("Values are different")
	}
}

func TestMapToStruct(t *testing.T) {
	s := NewMT_Source()

	gen, err := CopyToNew(s, reflect.TypeOf(&ST_Source{}))
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.(*ST_Source)
	if !issd {
		t.Fatal("Result is not *ST_Source")
	}

	if s["String1"].(string) != sd.String1 ||
		s["Int1"].(int) != sd.Int1 ||
		s["Float1"].(float32) != sd.Float1 {
		t.Fatal("Values are different")
	}
}

func TestMapToSlice(t *testing.T) {
	s := map[int]interface{}{
		0: "first",
		3: "third",
	}

	gen, err := CopyToNew(s, reflect.TypeOf([]interface{}{}))
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.([]interface{})
	if !issd {
		t.Fatal("Result is not []interface{}")
	}

	if len(sd) != 4 ||
		sd[0].(string) != s[0].(string) ||
		sd[3].(string) != s[3].(string) {
		t.Fatal("Values are different")
	}
}

func TestMapToSliceStringKey(t *testing.T) {
	s := map[string]interface{}{
		"0": "first",
		"3": "third",
	}

	gen, err := CopyToNew(s, reflect.TypeOf([]interface{}{}))
	if err != nil {
		t.Fatal(err)
	}

	sd, issd := gen.([]interface{})
	if !issd {
		t.Fatal("Result is not []interface{}")
	}

	if len(sd) != 4 ||
		sd[0].(string) != s["0"].(string) ||
		sd[3].(string) != s["3"].(string) {
		t.Fatal("Values are different")
	}
}
