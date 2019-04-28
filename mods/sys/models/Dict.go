package models

import "github.com/kuuland/kuu"

// Dict 系统字典
type Dict struct {
	ID        string      `json:"_id" displayName:"系统字典"`
	Code      string      `name:"字典编码"`
	Name      string      `name:"字典名称"`
	Values    []DictValue `name:"字典值"`
	IsBuiltIn bool        `name:"是否系统内置"`
	// 标准字段
	CreatedBy interface{} `name:"创建人" join:"User<Username,Name>"`
	CreatedAt int64       `name:"创建时间"`
	UpdatedBy interface{} `name:"修改人" join:"User<Username,Name>"`
	UpdatedAt int64       `name:"修改时间"`
	IsDeleted bool        `name:"是否已删除"`
	Remark    string      `name:"备注"`
}

// DictValue 字典值
type DictValue struct {
	Label string `name:"字典标签"`
	Value string `name:"字典键值"`
	Sort  int    `name:"排序号"`
}

func getDict(cond interface{}) (v *Dict) {
	Dict := kuu.Model("Dict")
	Dict.One(cond, v)
	return
}

// GetDictByValue 根据字典编码和值查询字典
func GetDictByValue(dictCode string, value string) *DictValue {
	dict := getDict(kuu.H{"Code": dictCode})
	if dict == nil {
		return nil
	}
	for _, item := range dict.Values {
		if item.Value == value {
			return &item
		}
	}
	return nil
}

// GetDictByLabel 根据字典编码和值标签查询字典
func GetDictByLabel(dictCode string, label string) *DictValue {
	dict := getDict(kuu.H{"Code": dictCode})
	if dict == nil {
		return nil
	}
	for _, item := range dict.Values {
		if item.Label == label {
			return &item
		}
	}
	return nil
}
