package goxcopy

import (
	"reflect"
	"strings"
)

func (c *Config) GetStructTagFields(field reflect.StructField) []string {
	var ret []string

	tag := field.Tag.Get(c.StructTagName)
	if tag != "" {
		tag_fields := strings.Split(tag, ",")
		if len(tag_fields) > 0 {
			ret = tag_fields
		}
	}

	return ret
}

func ReverseStrSlice(str []string) []string {
	var ret []string
	for i := len(str) - 1; i >= 0; i-- {
		ret = append(ret, str[i])
	}
	return ret
}
