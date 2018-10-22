package db

import (
	"math"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
)

const (
	// ALL 全量模式
	ALL = "ALL"
	// PAGE 分页模式
	PAGE = "PAGE"
)

// Params 定义了查询参数常用结构
type Params struct {
	ID      string
	Page    int
	Size    int
	Range   string
	Sort    []string
	Project map[string]int
	Cond    map[string]interface{}
}

// IModel 定义了模型持久化统一操作接口
type IModel interface {
	Create(interface{}) error
	Remove(kuu.H) error
	RemoveAll(kuu.H) (interface{}, error)
	Update(kuu.H, kuu.H) error
	UpdateAll(kuu.H, kuu.H) (interface{}, error)
	List(*Params, interface{}) (kuu.H, error)
	ID(*Params, interface{}) error
}

// Model 基于Mongo的模型操作实现
type Model struct {
	Name      string
	QueryHook func(query *mgo.Query)
}

// Create 实现新增
func (m *Model) Create(docs interface{}) error {
	C := C(m.Name)
	defer C.Database.Session.Close()
	return C.Insert(docs)
}

// Remove 实现删除
func (m *Model) Remove(cond kuu.H) error {
	C := C(m.Name)
	defer C.Database.Session.Close()
	return C.Remove(cond)
}

// RemoveAll 实现删除
func (m *Model) RemoveAll(cond kuu.H) (interface{}, error) {
	C := C(m.Name)
	defer C.Database.Session.Close()
	return C.RemoveAll(cond)
}

// Update 实现更新
func (m *Model) Update(cond kuu.H, doc kuu.H) error {
	C := C(m.Name)
	defer C.Database.Session.Close()
	return C.Update(cond, doc)
}

// UpdateAll 实现更新
func (m *Model) UpdateAll(cond kuu.H, doc kuu.H) (interface{}, error) {
	C := C(m.Name)
	defer C.Database.Session.Close()
	return C.UpdateAll(cond, doc)
}

// List 实现列表查询
func (m *Model) List(p *Params, list interface{}) (kuu.H, error) {
	C := C(m.Name)
	defer C.Database.Session.Close()
	query := C.Find(p.Cond)
	totalrecords, err := query.Count()
	if err != nil {
		return nil, err
	}
	if p.Project != nil {
		query.Select(p.Project)
	}
	if p.Range == PAGE {
		query.Skip((p.Page - 1) * p.Size).Limit(p.Size)
	}
	query.Sort(p.Sort...)
	if m.QueryHook != nil {
		m.QueryHook(query)
	}
	if err := query.All(list); err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]kuu.H, 0)
	}
	data := kuu.H{
		"list":         list,
		"totalrecords": totalrecords,
	}
	if p.Range == PAGE {
		totalpages := int(math.Ceil(float64(totalrecords) / float64(p.Size)))
		data["totalpages"] = totalpages
		data["page"] = p.Page
		data["size"] = p.Size
	}
	if p.Sort != nil && len(p.Sort) > 0 {
		data["sort"] = p.Sort
	}
	if p.Project != nil {
		data["project"] = p.Project
	}
	if p.Cond != nil {
		data["cond"] = p.Cond
	}
	if p.Range != "" {
		data["range"] = p.Range
	}
	return data, nil
}

// ID 实现ID查询
func (m *Model) ID(p *Params, data interface{}) error {
	C := C(m.Name)
	defer C.Database.Session.Close()
	id := p.ID
	query := C.FindId(bson.ObjectIdHex(id))
	if p.Project != nil {
		query.Select(p.Project)
	}
	if m.QueryHook != nil {
		m.QueryHook(query)
	}
	return query.One(data)
}
