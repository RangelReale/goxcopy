package goxcopy

type FieldMap struct {
	Fieldname *string
}

func NewFieldMap() *FieldMap {
	return &FieldMap{}
}

func (f *FieldMap) SetFieldname(fieldname string) *FieldMap {
	f.Fieldname = &fieldname
	return f
}
