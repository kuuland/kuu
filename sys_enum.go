package kuu

import "sync"

var (
	enumMap   = make(map[string]*EnumDesc)
	enumMapMu sync.RWMutex
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

// EnumMap
func EnumMap() map[string]*EnumDesc {
	enumMapMu.RLock()
	defer enumMapMu.RUnlock()
	return enumMap
}

// EnumList
func EnumList() (list []*EnumDesc) {
	for _, item := range EnumMap() {
		list = append(list, item)
	}
	return
}

// Enum
func Enum(classCode string, className ...string) *EnumDesc {
	enumMapMu.Lock()
	defer enumMapMu.Unlock()

	if v, has := enumMap[classCode]; has {
		return v
	}
	desc := &EnumDesc{
		ClassCode: classCode,
		Values:    make(map[interface{}]string),
	}
	if len(className) > 0 && className[0] != "" {
		desc.ClassName = className[0]
	}
	enumMap[desc.ClassCode] = desc
	return desc
}

// GetEnumItem
func GetEnumItem(classCode string) *EnumDesc {
	enumMapMu.RLock()
	defer enumMapMu.RUnlock()

	return enumMap[classCode]
}

// GetEnumLabel
func GetEnumLabel(classCode string, value interface{}) (label string) {
	item := GetEnumItem(classCode)
	if item != nil {
		label = item.Values[value]
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
