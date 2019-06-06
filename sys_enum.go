package kuu

var (
	enumMap  = make(map[string]*EnumDesc)
	enumList = make([]*EnumDesc, 0)
)

// EnumDesc
type EnumDesc struct {
	ClassCode string
	ClassName string
	Values    map[interface{}]string
}

// EnumItem
type EnumItem struct {
	Label string
	Value interface{}
}

// EnumList
func EnumList() []*EnumDesc {
	return enumList
}

// Enum
func Enum(classCode string, className ...string) *EnumDesc {
	desc := &EnumDesc{ClassCode: classCode, Values: make(map[interface{}]string)}
	if len(className) > 0 && className[0] != "" {
		desc.ClassName = className[0]
	}
	enumMap[desc.ClassCode] = desc
	enumList = append(enumList, desc)
	return desc
}

// Add
func (d *EnumDesc) Add(value interface{}, label ...string) *EnumDesc {
	if len(label) > 0 {
		d.Values[value] = label[0]
	} else {
		d.Values[value] = "-"
	}
	return d
}
