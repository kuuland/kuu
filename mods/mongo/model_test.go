package mongo

import (
	"fmt"
	"testing"
	"time"

	"github.com/kuuland/kuu"
)

type Dict struct {
	ID        string      `json:"_id" displayName:"系统字典"`
	Code      string      `name:"字典编码"`
	Name      string      `name:"字典名称"`
	Values    []DictValue `name:"字典值" array:"true" join:"DictValue"`
	IsBuiltIn bool        `name:"是否系统内置"`
	// 标准字段
	CreatedBy interface{} `name:"创建人" join:"User<Username,Name>"`
	CreatedAt time.Time   `name:"创建时间"`
	UpdatedBy interface{} `name:"创建人" join:"User<Username,Name>"`
	UpdatedAt time.Time   `name:"修改时间"`
	IsDeleted bool        `name:"是否已删除"`
	Remark    string      `name:"备注"`
}

type DictValue struct {
	ID        string `json:"_id" displayName:"字典值"`
	Dict      Dict   `name:"所属字典" join:"Dict<Code,Name>"`
	Label     string `name:"字典标签"`
	Value     string `name:"字典键值"`
	Sort      int    `name:"排序号"`
	IsBuiltIn bool   `name:"是否系统内置"`
	// 标准字段
	CreatedBy interface{} `name:"创建人" join:"User<Username,Name>"`
	CreatedAt time.Time   `name:"创建时间"`
	UpdatedBy interface{} `name:"创建人" join:"User<Username,Name>"`
	UpdatedAt time.Time   `name:"修改时间"`
	IsDeleted bool        `name:"是否已删除"`
	Remark    string      `name:"备注"`
}

func init() {
	kuu.Import(All())
	kuu.RegisterModel(&Dict{}, &DictValue{})
	kuu.New(kuu.H{
		"mongo": "mongodb://root:kuuland@127.0.0.1:27017/kuu?authSource=admin&maxPoolSize=50",
	})
}
func TestJoinCreate(t *testing.T) {
	dict := &Dict{
		Code:      "dict1",
		Name:      "字典1",
		IsBuiltIn: true,
		Values: []DictValue{
			DictValue{
				Label: "标签11",
				Value: "11",
			},
			DictValue{
				Label: "标签12",
				Value: "12",
			},
		},
	}
	DictModel := kuu.Model("Dict")
	if ret, err := DictModel.Create(dict); err != nil {
		kuu.Error(err)
	} else {
		fmt.Println(ret)
	}
}

func TestCheckID(t *testing.T) {
	before := kuu.H{
		"_id": kuu.H{
			"$in": []string{
				"5c4fd138b0fd26fddb68499b",
			},
		},
	}
	after := checkID(before)
	fmt.Println(before)
	fmt.Println(after)
}
