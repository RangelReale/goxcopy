package goxcopy

import "testing"

func TestPrimitiveTypedInteger(t *testing.T) {
	type PT_TI1 int
	type PT_TI2 int

	const (
		TI1_FIRST  PT_TI1 = 0
		TI1_SECOND        = 1
	)

	const (
		TI2_FIRST  PT_TI2 = 0
		TI2_SECOND        = 1
	)

	v1 := TI1_SECOND

	var v2 PT_TI2

	err := Copy(v1, &v2)
	if err != nil {
		t.Fatal(err)
	}

	if v2 != TI2_SECOND {
		t.Fatalf("Value should be TI2_SECOND, but is %v", v2)
	}

}
