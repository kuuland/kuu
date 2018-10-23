package mongo

import (
	"fmt"
	"testing"

	"github.com/kuuland/kuu"
)

var uri = "mongodb://root:kuuland@127.0.0.1:27017/kuu?authSource=admin&maxPoolSize=50"

func TestConnect(t *testing.T) {
	Connect(uri)
	if n, err := C("user").Count(); err == nil {
		fmt.Println(n)
	} else {
		fmt.Println(err)
	}
}

func ExampleModel_list() {
	User := Model("User")
	// 默认分页
	var userList []kuu.H
	User.List(&Params{}, userList)
	// 带参查询
	User.List(&Params{
		Page: 2,
		Size: 50,
		Sort: []string{"-CreatedAt"},
		Project: map[string]int{
			"Password": -1,
		},
	}, &userList)
	// 全量查询
	User.List(&Params{
		Range: ALL,
		Sort:  []string{"-CreatedAt"},
		Project: map[string]int{
			"Password": -1,
		},
	}, userList)
}
func ExampleModel_one() {
	User := Model("User")
	var userKuu kuu.H
	User.One(&Params{
		Cond: kuu.H{
			"Username": "kuu",
			"Password": "123456",
		},
	}, &userKuu)
}

func ExampleModel_id() {
	User := Model("User")
	var userKuu kuu.H
	User.ID(&Params{
		ID: "5bc0865cb7c851165e6bbac0",
	}, &userKuu)
}

func ExampleModel_create() {
	User := Model("User")
	// 单个新增
	User.Create(kuu.H{
		"Password": "123456",
		"Username": "kuu",
	})
	// 批量新增
	User.Create(kuu.H{
		"Password": "123456",
		"Username": "kuu1",
	}, kuu.H{
		"Password": "123456",
		"Username": "kuu2",
	})
	docs := []interface{}{
		kuu.H{
			"Password": "123456",
			"Username": "kuu3",
		},
		kuu.H{
			"Password": "123456",
			"Username": "kuu4",
		},
	}
	User.Create(docs...)
}
func ExampleModel_update() {
	User := Model("User")
	// 单个修改
	User.Update(kuu.H{
		"Username": "kuu",
	}, kuu.H{
		"Password": "654321",
	})
	// 批量修改
	User.UpdateAll(kuu.H{
		"Username": "kuu",
	}, kuu.H{
		"Password": "654321",
	})
}

func ExampleModel_remove() {
	User := Model("User")
	// 单个删除（逻辑删除）
	User.Remove(kuu.H{
		"Username": "kuu",
	})
	// 批量删除（逻辑删除）
	User.RemoveAll(kuu.H{
		"Username": "kuu",
	})
}
func ExampleModel_phyRemove() {
	User := Model("User")
	// 单个删除（物理删除）
	User.PhyRemove(kuu.H{
		"Username": "kuu",
	})
	// 批量删除（物理删除）
	User.PhyRemoveAll(kuu.H{
		"Username": "kuu",
	})
}
