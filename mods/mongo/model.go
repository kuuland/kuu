package mongo

import (
	"math"
	"time"

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
	Cond    kuu.H
}

// IModel 定义了模型持久化统一操作接口
type IModel interface {
	Create(...interface{}) ([]interface{}, error)
	Remove(kuu.H, kuu.H) error
	RemoveAll(kuu.H, kuu.H) (interface{}, error)
	PhyRemove(kuu.H) error
	PhyRemoveAll(kuu.H) (interface{}, error)
	Update(kuu.H, kuu.H) error
	UpdateAll(kuu.H, kuu.H) (interface{}, error)
	List(*Params, interface{}) (kuu.H, error)
	One(*Params, interface{}) (kuu.H, error)
	ID(*Params, interface{}) error
}

// Model 基于Mongo的模型操作实现
type Model struct {
	Name      string
	QueryHook func(query *mgo.Query)
	Session   *mgo.Session
}

// Create 实现新增
func (m *Model) Create(docs ...interface{}) ([]interface{}, error) {
	for index, item := range docs {
		var doc kuu.H
		kuu.JSONConvert(item, &doc)
		doc["_id"] = bson.NewObjectId()
		doc["CreatedAt"] = time.Now()
		docs[index] = doc
	}
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	err := C.Insert(docs...)
	return docs, err
}

// Remove 实现逻辑删除
func (m *Model) Remove(cond kuu.H, data ...kuu.H) error {
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	var doc kuu.H
	if len(data) > 0 {
		doc = data[0]
	}
	if doc == nil {
		doc = make(kuu.H)
	}
	doc["IsDeleted"] = true
	doc["UpdatedAt"] = time.Now()
	return C.Update(cond, doc)
}

// RemoveAll 实现逻辑删除
func (m *Model) RemoveAll(cond kuu.H, data ...kuu.H) (interface{}, error) {
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	var doc kuu.H
	if len(data) > 0 {
		doc = data[0]
	}
	if doc == nil {
		doc = make(kuu.H)
	}
	doc["IsDeleted"] = true
	doc["UpdatedAt"] = time.Now()
	return C.UpdateAll(cond, doc)
}

// PhyRemove 实现物理删除
func (m *Model) PhyRemove(cond kuu.H) error {
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	return C.Remove(cond)
}

// PhyRemoveAll 实现物理删除
func (m *Model) PhyRemoveAll(cond kuu.H) (interface{}, error) {
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	return C.RemoveAll(cond)
}

// Update 实现更新
func (m *Model) Update(cond kuu.H, doc kuu.H) error {
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	doc["UpdatedAt"] = time.Now()
	return C.Update(cond, doc)
}

// UpdateAll 实现更新
func (m *Model) UpdateAll(cond kuu.H, doc kuu.H) (interface{}, error) {
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	doc["UpdatedAt"] = time.Now()
	return C.UpdateAll(cond, doc)
}

// List 实现列表查询
func (m *Model) List(p *Params, list interface{}) (kuu.H, error) {
	// 参数加工
	if list == nil {
		list = make([]kuu.H, 0)
	}
	isDeleted := kuu.H{
		"$ne": true,
	}
	if p.Cond == nil {
		p.Cond = make(kuu.H)
	}
	if p.Cond["$and"] != nil {
		var and []kuu.H
		kuu.JSONConvert(p.Cond["$and"], &and)
		hasDr := false
		for _, item := range and {
			if item["IsDeleted"] != nil {
				hasDr = true
				break
			}
		}
		if !hasDr {
			and = append(and, kuu.H{
				"IsDeleted": isDeleted,
			})
			p.Cond["$and"] = and
		}
	} else {
		if p.Cond["IsDeleted"] == nil {
			p.Cond["IsDeleted"] = isDeleted
		}
	}

	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	query := C.Find(p.Cond)
	totalRecords, err := query.Count()
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
	result := []kuu.H{}
	if err := query.All(&result); err != nil {
		return nil, err
	}
	kuu.JSONConvert(result, list)
	data := kuu.H{
		"list":         list,
		"totalrecords": totalRecords,
	}
	if p.Range == PAGE {
		totalpages := int(math.Ceil(float64(totalRecords) / float64(p.Size)))
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
	if p.Cond == nil {
		p.Cond = make(kuu.H)
	}
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	id := p.ID
	query := C.FindId(bson.ObjectIdHex(id))
	if p.Project != nil {
		query.Select(p.Project)
	}
	if m.QueryHook != nil {
		m.QueryHook(query)
	}
	result := kuu.H{}
	err := query.One(result)
	if err == nil {
		kuu.JSONConvert(result, data)
	}
	return err
}

// One 实现单个查询
func (m *Model) One(p *Params, data interface{}) error {
	if p.Cond == nil {
		p.Cond = make(kuu.H)
	}
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	query := C.Find(p.Cond)
	if p.Project != nil {
		query.Select(p.Project)
	}
	if m.QueryHook != nil {
		m.QueryHook(query)
	}
	result := kuu.H{}
	err := query.One(result)
	if err == nil {
		kuu.JSONConvert(result, data)
	}
	return err
}
