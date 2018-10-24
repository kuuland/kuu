package mongo

import (
	"errors"
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

// Model 基于Mongo的模型操作实现
type Model struct {
	Name      string
	QueryHook func(query *mgo.Query)
	Session   *mgo.Session
}

// Create 实现新增（支持传入单个或者数组）
func (m *Model) Create(data interface{}) ([]interface{}, error) {
	docs := []interface{}{}
	if kuu.IsArray(data) {
		kuu.JSONConvert(data, &docs)
	} else {
		doc := kuu.H{}
		kuu.JSONConvert(data, &doc)
		docs = append(docs, doc)
	}
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

// Remove 实现基于条件的逻辑删除
func (m *Model) Remove(cond kuu.H, data ...kuu.H) error {
	var doc kuu.H
	if len(data) > 0 {
		doc = data[0]
	}
	_, err := m.remove(cond, doc, false)
	return err
}

// RemoveEntity 实现基于实体的逻辑删除
func (m *Model) RemoveEntity(entity interface{}, data ...kuu.H) error {
	var doc kuu.H
	if len(data) > 0 {
		doc = data[0]
	}
	var obj kuu.H
	kuu.JSONConvert(entity, &obj)
	if obj == nil || obj["_id"] == nil {
		return errors.New("_id is required")
	}
	cond := kuu.H{
		"_id": obj["_id"],
	}
	_, err := m.remove(cond, doc, false)
	return err
}

// RemoveAll 实现基于条件的批量逻辑删除
func (m *Model) RemoveAll(cond kuu.H, data ...kuu.H) (interface{}, error) {
	var doc kuu.H
	if len(data) > 0 {
		doc = data[0]
	}
	return m.remove(cond, doc, true)
}

// PhyRemove 实现基于条件的物理删除
func (m *Model) PhyRemove(cond kuu.H) error {
	_, err := m.phyRemove(cond, false)
	return err
}

// PhyRemoveEntity 实现基于实体的物理删除
func (m *Model) PhyRemoveEntity(entity interface{}) error {
	var obj kuu.H
	kuu.JSONConvert(entity, &obj)
	if obj == nil || obj["_id"] == nil {
		return errors.New("_id is required")
	}
	cond := kuu.H{
		"_id": obj["_id"],
	}
	_, err := m.phyRemove(cond, false)
	return err
}

// PhyRemoveAll 实现基于条件的批量物理删除
func (m *Model) PhyRemoveAll(cond kuu.H) (interface{}, error) {
	return m.phyRemove(cond, true)
}

// Update 实现基于条件的更新
func (m *Model) Update(cond kuu.H, doc kuu.H) error {
	_, err := m.update(cond, doc, false)
	return err
}

// UpdateEntity 实现基于实体的更新
func (m *Model) UpdateEntity(entity interface{}) error {
	var doc kuu.H
	kuu.JSONConvert(entity, &doc)
	if doc == nil || doc["_id"] == nil {
		return errors.New("_id is required")
	}
	cond := kuu.H{
		"_id": doc["_id"],
	}
	delete(doc, "_id")
	_, err := m.update(cond, doc, false)
	return err
}

// UpdateAll 实现基于条件的批量更新
func (m *Model) UpdateAll(cond kuu.H, doc kuu.H) (interface{}, error) {
	return m.update(cond, doc, true)
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
func (m *Model) ID(v interface{}, data interface{}) error {
	p := &Params{}
	switch v.(type) {
	case *Params:
		p = v.(*Params)
	case string:
		p = &Params{
			ID: v.(string),
		}
	}
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

func (m *Model) remove(cond kuu.H, doc kuu.H, all bool) (interface{}, error) {
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	if doc == nil {
		doc = make(kuu.H)
	}
	doc["IsDeleted"] = true
	doc["UpdatedAt"] = time.Now()
	cond = checkObjectID(cond)
	doc = checkUpdateSet(doc)
	if all {
		return C.UpdateAll(cond, doc)
	}
	return nil, C.Update(cond, doc)
}

func (m *Model) phyRemove(cond kuu.H, all bool) (interface{}, error) {
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	cond = checkObjectID(cond)
	if all {
		return C.RemoveAll(cond)
	}
	return nil, C.Remove(cond)
}

func (m *Model) update(cond kuu.H, doc kuu.H, all bool) (interface{}, error) {
	C := C(m.Name)
	m.Session = C.Database.Session
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
	}()
	doc = setUpdateValue(doc, "UpdatedAt", time.Now())
	cond = checkObjectID(cond)
	doc = checkUpdateSet(doc)
	if all {
		return C.UpdateAll(cond, doc)
	}
	return nil, C.Update(cond, doc)
}

func setUpdateValue(doc kuu.H, key string, value interface{}) kuu.H {
	if doc["$set"] != nil {
		set := kuu.H{}
		kuu.JSONConvert(doc["$set"], &set)
		set[key] = value
		doc["$set"] = set
	} else {
		doc[key] = value
	}
	return doc
}

func checkObjectID(cond kuu.H) kuu.H {
	if cond["_id"] != nil {
		cond["_id"] = bson.ObjectIdHex(cond["_id"].(string))
	}
	return cond
}

func checkUpdateSet(doc kuu.H) kuu.H {
	if doc["$set"] == nil {
		doc = kuu.H{
			"$set": doc,
		}
	}
	return doc
}
