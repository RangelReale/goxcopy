package goxcopy

import (
	"reflect"
	"testing"
)

func TestMergeMap(t *testing.T) {

	s1 := map[string]string{
		"value1": "s1_value1",
		"value2": "s1_value2",
	}

	s2 := map[string]string{
		"value3": "s2_value3",
		"value4": "s2_value4",
	}

	s3 := map[string]string{
		"value1": "s3_value1",
		"value4": "s3_value4",
		"value5": "s3_value5",
	}

	ret, err := MergeToNew(reflect.TypeOf(map[string]string{}), s1, s2, s3)
	if err != nil {
		t.Fatal(err)
	}

	retval, ok := ret.(map[string]string)
	if !ok {
		t.Fatal("Result isn't map[string]string")
	}

	if len(retval) != 5 || retval["value1"] != "s3_value1" || retval["value2"] != "s1_value2" ||
		retval["value3"] != "s2_value3" || retval["value4"] != "s3_value4" || retval["value5"] != "s3_value5" {
		t.Fatal("Result values are not the expected ones")
	}
}
