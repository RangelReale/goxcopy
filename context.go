package goxcopy

import (
	"github.com/RangelReale/rprim"
	"reflect"
	"strings"
)

type Context struct {
	Fields []reflect.Value
}

func NewContext() *Context {
	return &Context{}
}

func (c *Context) Dup() *Context {
	ret := &Context{}
	for _, f := range c.Fields {
		ret.Fields = append(ret.Fields, f)
	}
	return ret
}

func (c *Context) PushField(fieldname reflect.Value) {
	c.Fields = append(c.Fields, fieldname)
}

func (c *Context) PopField() {
	if len(c.Fields) > 0 {
		c.Fields = c.Fields[:len(c.Fields)-1]
	}
}

func (c *Context) FieldsAsStringSlice() []string {
	var ret []string
	for _, f := range c.Fields {
		s, err := rprim.ConvertToString(f)
		if err != nil {
			ret = append(ret, "<unknown>")
		} else {
			ret = append(ret, s)
		}
	}
	return ReverseStrSlice(ret)
}

func (c *Context) FieldsAsString() string {
	return strings.Join(c.FieldsAsStringSlice(), ".")
}
