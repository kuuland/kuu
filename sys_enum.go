package kuu

var (
	enumMap  = make(map[string]*EnumDesc)
	enumList = make([]*EnumDesc, 0)
)

// EnumDesc
type EnumDesc struct {
	ClassCode string
	ClassName string
	Values    map[interface{}]string `json:"-"`
	Items     []EnumItem             `json:"Values"`
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

// GetEnumItem
func GetEnumItem(classCode string) *EnumDesc {
	return enumMap[classCode]
}

// GetEnumLabel
func GetEnumLabel(classCode string, value interface{}) (label string) {
	if v, has := enumMap[classCode]; has && v != nil {
		label = v.Values[value]
	}
	return
}

// Add
func (d *EnumDesc) Add(value interface{}, label ...string) *EnumDesc {
	if len(label) > 0 {
		d.Values[value] = label[0]
		d.Items = append(d.Items, EnumItem{Value: value, Label: label[0]})
	} else {
		d.Values[value] = "-"
		d.Items = append(d.Items, EnumItem{Value: value})
	}
	return d
}
