package kuu

var EnumKey = BuildKey("EMUM")

type EnumDesc struct {
	ClassCode string
	ClassName string
	Values    map[interface{}]string `json:"-"`
	Items     []EnumItem             `json:"Values"`
}

type EnumItem struct {
	Label string
	Value interface{}
}

func EnumMap() map[string]*EnumDesc {
	rawMap := DefaultCache.HGetAll(EnumKey)
	var m = make(map[string]*EnumDesc)
	for k, v := range rawMap {
		var item EnumDesc
		err := JSONParse(v, &item)
		if err != nil {
			continue
		}
		item.Values = make(map[interface{}]string)
		for _, enumItem := range item.Items {
			item.Values[enumItem.Value] = enumItem.Label
		}
		m[k] = &item
	}
	return m
}

func EnumList() (list []*EnumDesc) {
	for _, item := range EnumMap() {
		list = append(list, item)
	}
	return
}

func Enum(classCode string, className ...string) *EnumDesc {
	desc := &EnumDesc{
		ClassCode: classCode,
		Values:    make(map[interface{}]string),
	}
	if len(className) > 0 && className[0] != "" {
		desc.ClassName = className[0]
	}
	return desc
}

func GetEnumItem(classCode string) *EnumDesc {
	raw := DefaultCache.HGet(EnumKey, classCode)
	var item EnumDesc
	err := JSONParse(raw, &item)
	if err != nil {
		return nil
	}
	return &item
}

func GetEnumLabel(classCode string, value interface{}) (label string) {
	item := GetEnumItem(classCode)
	if item != nil {
		label = item.Values[value]
	}
	return
}

func (d *EnumDesc) Add(value interface{}, label ...string) *EnumDesc {
	if len(label) > 0 {
		d.Values[value] = label[0]
		d.Items = append(d.Items, EnumItem{Value: value, Label: label[0]})
	} else {
		d.Values[value] = "-"
		d.Items = append(d.Items, EnumItem{Value: value})
	}
	DefaultCache.HSet(EnumKey, d.ClassCode, JSONStringify(d, true))
	return d
}
