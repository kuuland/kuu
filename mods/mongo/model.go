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
	schema     *kuu.Schema
	Scope      *Scope
	Collection string
	Name       string
	Session    *mgo.Session
}

// New 实现New接口
func (m *Model) New(schema *kuu.Schema) kuu.IModel {
	n := &Model{
		schema:     schema,
		Name:       schema.Name,
		Collection: schema.Collection,
	}
	return n
}

// Schema 实现Schema接口
func (m *Model) Schema() *kuu.Schema {
	return m.schema
}

// Create 实现新增（支持传入单个或者数组）
func (m *Model) Create(data interface{}) ([]interface{}, error) {
	m.Scope = &Scope{
		Operation: "Create",
		Cache:     make(kuu.H),
	}
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
		if doc["CreatedBy"] != nil {
			switch doc["CreatedBy"].(type) {
			case string:
				doc["CreatedBy"] = bson.ObjectIdHex(doc["CreatedBy"].(string))
			case bson.ObjectId:
				doc["CreatedBy"] = doc["CreatedBy"].(bson.ObjectId)
			}
		}
		// 设置UpdatedXx初始值等于CreatedXx
		doc["UpdatedAt"] = doc["CreatedAt"]
		doc["UpdatedBy"] = doc["CreatedBy"]
		docs[index] = doc
	}
	C := C(m.Collection)
	m.Session = C.Database.Session
	m.Scope.Session = m.Session
	m.Scope.Collection = C
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
		m.Scope = nil
	}()
	m.Scope.CallMethod(BeforeSaveEnum, m.schema)
	m.Scope.CallMethod(BeforeCreateEnum, m.schema)
	// 先保存外键
	handleJoinBeforeCreate(docs, m.schema)
	err := C.Insert(docs...)
	m.Scope.CallMethod(AfterCreateEnum, m.schema)
	m.Scope.CallMethod(AfterSaveEnum, m.schema)
	return docs, err
}

func handleJoinBeforeCreate(docs []interface{}, schema *kuu.Schema) []interface{} {
	for index, item := range docs {
		var doc kuu.H
		kuu.JSONConvert(item, &doc)
		// 找出引用字段
		joinFields := getJoinFields(schema, nil)
		for _, field := range joinFields {
			refData := doc[field.Code]
			if field.IsArray {
				// 数组
				if v, ok := refData.([]string); ok {
					arr := []interface{}{}
					for _, str := range v {
						arr = append(arr, bson.ObjectIdHex(str))
					}
					doc[field.Code] = arr
				} else if v, ok := refData.([]bson.ObjectId); ok {
					doc[field.Code] = v
				} else {
					var refDocs []kuu.H
					kuu.JSONConvert(doc[field.Code], &refDocs)
					arr := []interface{}{}
					newDocs := []kuu.H{}
					for _, refDoc := range refDocs {
						if refDoc["_id"] != nil {
							if v, ok := refDoc["_id"].(string); ok {
								arr = append(arr, bson.ObjectIdHex(v))
							} else if v, ok := refDoc["_id"].(bson.ObjectId); ok {
								arr = append(arr, v)
							}
						} else {
							newDocs = append(newDocs, refDoc)
						}
					}
					if len(newDocs) > 0 {
						RefModel := kuu.Model(field.JoinName)
						if ret, err := RefModel.Create(newDocs); err == nil {
							var retDocs []kuu.H
							kuu.JSONConvert(ret, retDocs)
							for _, retDoc := range retDocs {
								arr = append(arr, retDoc["_id"])
							}
						}
					}
					doc[field.Code] = arr
				}
			} else {
				// 单个
				if v, ok := refData.(string); ok {
					doc[field.Code] = bson.ObjectIdHex(v)
				} else if v, ok := refData.(bson.ObjectId); ok {
					doc[field.Code] = v
				} else {
					var refDoc kuu.H
					kuu.JSONConvert(doc[field.Code], &refDoc)
					if refDoc == nil {
						continue
					}
					if refDoc["_id"] == nil {
						RefModel := kuu.Model(field.JoinName)
						if ret, err := RefModel.Create(refDoc); err == nil {
							var newDoc kuu.H
							kuu.JSONConvert(ret[0], newDoc)
							doc[field.Code] = newDoc["_id"]
						}
					} else {
						if v, ok := refDoc["_id"].(string); ok {
							doc[field.Code] = bson.ObjectIdHex(v)
						} else if v, ok := refDoc["_id"].(bson.ObjectId); ok {
							doc[field.Code] = v
						}
					}
				}
			}
		}
		docs[index] = doc
	}
	return docs
}

// Remove 实现基于条件的逻辑删除
func (m *Model) Remove(selector interface{}) error {
	return m.RemoveWithData(selector, nil)
}

// RemoveWithData 实现基于条件的逻辑删除
func (m *Model) RemoveWithData(selector interface{}, data interface{}) error {
	var (
		cond kuu.H
		doc  kuu.H
	)
	kuu.JSONConvert(selector, &cond)
	kuu.JSONConvert(data, &doc)
	_, err := m.remove(cond, doc, false)
	return err
}

// RemoveEntity 实现基于实体的逻辑删除
func (m *Model) RemoveEntity(entity interface{}) error {
	return m.RemoveEntityWithData(entity, nil)
}

// RemoveEntityWithData 实现基于实体的逻辑删除
func (m *Model) RemoveEntityWithData(entity interface{}, data interface{}) error {
	var (
		doc kuu.H
		obj kuu.H
	)
	kuu.JSONConvert(entity, &obj)
	kuu.JSONConvert(data, &doc)
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
func (m *Model) RemoveAll(selector interface{}) (interface{}, error) {
	return m.RemoveAllWithData(selector, nil)
}

// RemoveAllWithData 实现基于条件的批量逻辑删除
func (m *Model) RemoveAllWithData(selector interface{}, data interface{}) (interface{}, error) {
	var (
		cond kuu.H
		doc  kuu.H
	)
	kuu.JSONConvert(selector, &cond)
	kuu.JSONConvert(data, &doc)
	return m.remove(cond, doc, true)
}

// PhyRemove 实现基于条件的物理删除
func (m *Model) PhyRemove(selector interface{}) error {
	var cond kuu.H
	kuu.JSONConvert(selector, &cond)
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
func (m *Model) PhyRemoveAll(selector interface{}) (interface{}, error) {
	var cond kuu.H
	kuu.JSONConvert(selector, &cond)
	return m.phyRemove(cond, true)
}

// Update 实现基于条件的更新
func (m *Model) Update(selector interface{}, data interface{}) error {
	var (
		cond kuu.H
		doc  kuu.H
	)
	kuu.JSONConvert(selector, &cond)
	kuu.JSONConvert(data, &doc)
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
func (m *Model) UpdateAll(selector interface{}, data interface{}) (interface{}, error) {
	var (
		cond kuu.H
		doc  kuu.H
	)
	kuu.JSONConvert(selector, &cond)
	kuu.JSONConvert(data, &doc)
	return m.update(cond, doc, true)
}

// List 实现列表查询
func (m *Model) List(a interface{}, list interface{}) (kuu.H, error) {
	m.Scope = &Scope{
		Operation: "List",
		Cache:     make(kuu.H),
	}
	p := &Params{}
	kuu.JSONConvert(a, p)
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
	if p.Cond["_id"] != nil {
		if v, ok := p.Cond["_id"].(string); ok {
			p.Cond["_id"] = bson.ObjectIdHex(v)
		}
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

	C := C(m.Collection)
	m.Session = C.Database.Session
	m.Scope.Session = m.Session
	m.Scope.Collection = C
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
		m.Scope = nil
	}()
	query := C.Find(p.Cond)
	m.Scope.Query = query
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
	m.Scope.CallMethod(BeforeFindEnum, m.schema)
	var result []kuu.H
	if err := query.All(&result); err != nil {
		return nil, err
	}
	listJoin(m.Scope.Session, m.schema, p.Project, result)
	kuu.JSONConvert(result, list)
	if list == nil {
		list = make([]kuu.H, 0)
	}
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
	m.Scope.CallMethod(AfterFindEnum, m.schema)
	return data, nil
}

// ID 实现ID查询（支持传入“string”、“bson.ObjectId”和“&mongo.Params”三种类型的数据）
func (m *Model) ID(id interface{}, data interface{}) error {
	m.Scope = &Scope{
		Operation: "ID",
		Cache:     make(kuu.H),
	}
	p := &Params{}
	switch id.(type) {
	case Params:
		p = id.(*Params)
	case bson.ObjectId:
		p = &Params{
			ID: id.(bson.ObjectId).Hex(),
		}
	case string:
		p = &Params{
			ID: id.(string),
		}
	}
	if p.Cond == nil {
		p.Cond = make(kuu.H)
	}
	if p.ID == "" {
		kuu.JSONConvert(id, p)
	}
	C := C(m.Collection)
	m.Session = C.Database.Session
	m.Scope.Session = m.Session
	m.Scope.Collection = C
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
		m.Scope = nil
	}()
	v := p.ID
	query := C.FindId(bson.ObjectIdHex(v))
	m.Scope.Query = query
	if p.Project != nil {
		query.Select(p.Project)
	}
	result := kuu.H{}
	m.Scope.CallMethod(BeforeFindEnum, m.schema)
	err := query.One(&result)
	if err == nil {
		oneJoin(m.Scope.Session, m.schema, p.Project, result)
		kuu.JSONConvert(&result, data)
	}
	m.Scope.CallMethod(AfterFindEnum, m.schema)
	return err
}

// One 实现单个查询
func (m *Model) One(a interface{}, data interface{}) error {
	m.Scope = &Scope{
		Operation: "One",
		Cache:     make(kuu.H),
	}
	p := &Params{}
	kuu.JSONConvert(a, p)
	if p.Cond == nil {
		p.Cond = make(kuu.H)
	}
	C := C(m.Collection)
	m.Session = C.Database.Session
	m.Scope.Session = m.Session
	m.Scope.Collection = C
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
		m.Scope = nil
	}()
	query := C.Find(p.Cond)
	m.Scope.Query = query
	if p.Project != nil {
		query.Select(p.Project)
	}
	result := kuu.H{}
	m.Scope.CallMethod(BeforeFindEnum, m.schema)
	err := query.One(&result)
	if err == nil {
		oneJoin(m.Scope.Session, m.schema, p.Project, result)
		kuu.JSONConvert(&result, data)
	}
	m.Scope.CallMethod(AfterFindEnum, m.schema)
	return err
}

func (m *Model) remove(cond kuu.H, doc kuu.H, all bool) (ret interface{}, err error) {
	m.Scope = &Scope{
		Operation: "Remove",
		Cache:     make(kuu.H),
	}
	C := C(m.Collection)
	m.Session = C.Database.Session
	m.Scope.Session = m.Session
	m.Scope.Collection = C
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
		m.Scope = nil
	}()
	if doc == nil {
		doc = make(kuu.H)
	}
	doc["IsDeleted"] = true
	doc["UpdatedAt"] = time.Now()
	if doc["UpdatedBy"] != nil {
		switch doc["UpdatedBy"].(type) {
		case string:
			doc["UpdatedBy"] = bson.ObjectIdHex(doc["UpdatedBy"].(string))
		case bson.ObjectId:
			doc["UpdatedBy"] = doc["UpdatedBy"].(bson.ObjectId)
		}
	}
	cond = checkID(cond)
	doc = checkUpdateSet(doc)
	m.Scope.CallMethod(BeforeRemoveEnum, m.schema)
	if all {
		ret, err = C.UpdateAll(cond, doc)
	}
	err = C.Update(cond, doc)
	m.Scope.CallMethod(AfterRemoveEnum, m.schema)
	return ret, err
}

func (m *Model) phyRemove(cond kuu.H, all bool) (ret interface{}, err error) {
	m.Scope = &Scope{
		Operation: "PhyRemove",
		Cache:     make(kuu.H),
	}
	C := C(m.Collection)
	m.Session = C.Database.Session
	m.Scope.Session = m.Session
	m.Scope.Collection = C
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
		m.Scope = nil
	}()
	cond = checkID(cond)
	m.Scope.CallMethod(BeforePhyRemoveEnum, m.schema)
	if all {
		ret, err = C.RemoveAll(cond)
	}
	err = C.Remove(cond)
	m.Scope.CallMethod(AfterPhyRemoveEnum, m.schema)
	return ret, err
}

func (m *Model) update(cond kuu.H, doc kuu.H, all bool) (ret interface{}, err error) {
	m.Scope = &Scope{
		Operation: "Update",
		Cache:     make(kuu.H),
	}
	C := C(m.Collection)
	m.Session = C.Database.Session
	m.Scope.Session = m.Session
	m.Scope.Collection = C
	defer func() {
		C.Database.Session.Close()
		m.Session = nil
		m.Scope = nil
	}()
	doc = setUpdateValue(doc, "UpdatedAt", time.Now())
	if doc["UpdatedBy"] != nil {
		switch doc["UpdatedBy"].(type) {
		case string:
			doc["UpdatedBy"] = bson.ObjectIdHex(doc["UpdatedBy"].(string))
		case bson.ObjectId:
			doc["UpdatedBy"] = doc["UpdatedBy"].(bson.ObjectId)
		}
	}
	cond = checkID(cond)
	doc = checkUpdateSet(doc)
	m.Scope.CallMethod(BeforeUpdateEnum, m.schema)
	if all {
		ret, err = C.UpdateAll(cond, doc)
	}
	err = C.Update(cond, doc)
	m.Scope.CallMethod(AfterUpdateEnum, m.schema)
	return ret, err
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

func checkID(cond kuu.H) kuu.H {
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

func listJoin(session *mgo.Session, schema *kuu.Schema, project map[string]int, list []kuu.H) {
	fields := getJoinFields(schema, project)
	for _, field := range fields {
		// 拼装查询条件
		arr := []bson.ObjectId{}
		for _, result := range list {
			rawData := result[field.Code]
			if rawData == nil {
				continue
			}
			if field.IsArray {
				var itemArr []bson.ObjectId
				switch rawData.(type) {
				case []string:
					for _, item := range rawData.([]string) {
						id := bson.ObjectIdHex(item)
						if id != "" {
							arr = append(arr, id)
							itemArr = append(itemArr, id)
						}
					}
				case []bson.ObjectId:
					for _, item := range rawData.([]bson.ObjectId) {
						id := item
						arr = append(arr, id)
						itemArr = append(itemArr, id)
					}
				case []interface{}:
					for _, item := range rawData.([]interface{}) {
						var id bson.ObjectId
						switch item.(type) {
						case bson.ObjectId:
							id = item.(bson.ObjectId)
						case string:
							id = bson.ObjectIdHex(item.(string))
						}
						if id != "" {
							arr = append(arr, id)
							itemArr = append(itemArr, id)
						}
					}
				}
				result[field.Code] = itemArr
			} else {
				var id bson.ObjectId
				switch rawData.(type) {
				case bson.ObjectId:
					id = rawData.(bson.ObjectId)
				case string:
					id = bson.ObjectIdHex(rawData.(string))
				}
				if id != "" {
					arr = append(arr, id)
					result[field.Code] = id
				}
			}
		}
		if len(arr) == 0 {
			continue
		}
		// 执行查询
		var ret []kuu.H
		findJoinData(session, field, kuu.H{
			"_id": kuu.H{
				"$in": arr,
			},
		}, &ret)
		// 逐条填充引用数据
		hexMap := getHexMap(ret)
		for _, result := range list {
			if result[field.Code] == nil {
				continue
			}
			if field.IsArray {
				if s, ok := (result[field.Code]).([]bson.ObjectId); ok {
					l := make([]interface{}, len(s))
					for i, item := range s {
						key := item.Hex()
						if key == "" {
							l[i] = item
						} else {
							l[i] = hexMap[key]
						}
					}
					result[field.Code] = l
				}
			} else {
				if s, ok := (result[field.Code]).(bson.ObjectId); ok {
					if s != "" {
						result[field.Code] = hexMap[s.Hex()]
					}
				}
			}
		}
	}
}

func oneJoin(session *mgo.Session, schema *kuu.Schema, project map[string]int, result kuu.H) {
	fields := getJoinFields(schema, project)
	for _, field := range fields {
		// 拼装查询条件
		rawData := result[field.Code]
		if rawData == nil {
			continue
		}
		arr := []bson.ObjectId{}
		if field.IsArray {
			var itemArr []bson.ObjectId
			switch rawData.(type) {
			case []string:
				for _, item := range rawData.([]string) {
					id := bson.ObjectIdHex(item)
					if id != "" {
						arr = append(arr, id)
						itemArr = append(itemArr, id)
					}
				}
			case []bson.ObjectId:
				for _, item := range rawData.([]bson.ObjectId) {
					id := item
					arr = append(arr, id)
					itemArr = append(itemArr, id)
				}
			case []interface{}:
				for _, item := range rawData.([]interface{}) {
					var id bson.ObjectId
					switch item.(type) {
					case bson.ObjectId:
						id = item.(bson.ObjectId)
					case string:
						id = bson.ObjectIdHex(item.(string))
					}
					if id != "" {
						arr = append(arr, id)
						itemArr = append(itemArr, id)
					}
				}
			}
			result[field.Code] = itemArr
		} else {
			var id bson.ObjectId
			switch rawData.(type) {
			case bson.ObjectId:
				id = rawData.(bson.ObjectId)
			case string:
				id = bson.ObjectIdHex(rawData.(string))
			}
			if id != "" {
				arr = append(arr, id)
				result[field.Code] = id
			}
		}
		// 执行查询
		var ret []kuu.H
		findJoinData(session, field, kuu.H{
			"_id": kuu.H{
				"$in": arr,
			},
		}, &ret)
		// 替换引用数据
		hexMap := getHexMap(ret)
		if field.IsArray {
			if s, ok := (result[field.Code]).([]bson.ObjectId); ok {
				l := make([]interface{}, len(s))
				for i, item := range s {
					key := item.Hex()
					if key == "" {
						l[i] = item
					} else {
						l[i] = hexMap[key]
					}
				}
				result[field.Code] = l
			}
		} else {
			if s, ok := (result[field.Code]).(bson.ObjectId); ok {
				if s != "" {
					result[field.Code] = hexMap[s.Hex()]
				}
			}
		}
		result[field.Code] = ret
	}
}

func getJoinFields(schema *kuu.Schema, project map[string]int) []*kuu.SchemaField {
	fields := []*kuu.SchemaField{}
	for _, field := range schema.Fields {
		if field.JoinName == "" {
			continue
		}
		if project == nil {
			fields = append(fields, field)
		} else if _, ok := project[field.Name]; ok {
			fields = append(fields, field)
		}
	}
	return fields
}

// 查询引用数据
func findJoinData(session *mgo.Session, field *kuu.SchemaField, selector interface{}, result *[]kuu.H) {
	if session == nil || field == nil {
		return
	}
	joinSchema := kuu.GetSchema(field.JoinName)
	C := C(joinSchema.Collection, session)
	query := C.Find(selector)
	if field.JoinSelect != nil {
		if s, ok := field.JoinSelect["_id"]; ok && s == 0 {
			delete(field.JoinSelect, "_id")
		}
		query.Select(field.JoinSelect)
	}
	query.All(result)
	return
}

func getHexMap(ret []kuu.H) kuu.H {
	hexMap := kuu.H{}
	for _, item := range ret {
		if s, ok := item["_id"].(bson.ObjectId); ok {
			hexMap[s.Hex()] = item
		}
	}
	return hexMap
}
