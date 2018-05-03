package goxcopy

import "testing"

func TestArrayToArray(t *testing.T) {
	src := [...]string{"1", "2", "3"}
	var dst [3]string

	err := CopyToExisting(src, &dst)
	if err != nil {
		t.Fatal(err)
	}

	if src[0] != dst[0] || src[1] != dst[1] || src[2] != dst[2] {
		t.Fatal("Values are different")
	}
}
