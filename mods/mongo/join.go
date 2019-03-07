package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
)

func getJoinFields(schema *kuu.Schema, project map[string]int) []kuu.SchemaField {
	fields := []kuu.SchemaField{}
	rawFields := *schema.Fields
	for _, field := range rawFields {
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

func ensureJoinDoc(doc kuu.H, schema *kuu.Schema) kuu.H {
	joinFields := getJoinFields(schema, nil)
	for _, field := range joinFields {
		refData := doc[field.Code]
		if refData == nil {
			continue
		}
		if field.IsArray {
			if v, ok := refData.([]string); ok {
				arr := []interface{}{}
				for _, str := range v {
					if str == "" {
						continue
					}
					arr = append(arr, bson.ObjectIdHex(str))
				}
				doc[field.Code] = arr
			} else if v, ok := refData.([]bson.ObjectId); ok {
				doc[field.Code] = v
			}
		} else {
			if v, ok := refData.(string); ok && v != "" {
				doc[field.Code] = bson.ObjectIdHex(v)
			} else if v, ok := refData.(bson.ObjectId); ok && v != "" {
				doc[field.Code] = v
			}
		}
	}
	return doc
}

func handleJoinBeforeQuery(cond kuu.H, schema *kuu.Schema) kuu.H {
	if cond["$and"] != nil {
		if v, ok := cond["$and"].([]interface{}); ok {
			for index, item := range v {
				var doc kuu.H
				kuu.JSONConvert(item, &doc)
				v[index] = ensureJoinDoc(doc, schema)
			}
			cond["$and"] = v
		} else if v, ok := cond["$and"].([]kuu.H); ok {
			for index, doc := range v {
				v[index] = ensureJoinDoc(doc, schema)
			}
			cond["$and"] = v
		}
	} else {
		cond = ensureJoinDoc(cond, schema)
	}
	return cond
}

func handleJoinBeforeSave(docs []interface{}, schema *kuu.Schema) []interface{} {
	for index, item := range docs {
		var doc kuu.H
		kuu.JSONConvert(item, &doc)
		// 确保_id类型为ObjectId
		if v, ok := doc["_id"].(string); ok && v != "" {
			doc["_id"] = bson.ObjectIdHex(v)
		} else if v, ok := doc["_id"].(bson.ObjectId); ok && v != "" {
			doc["_id"] = v
		} else {
			delete(doc, "_id")
		}
		// 找出引用字段
		joinFields := getJoinFields(schema, nil)
		for _, field := range joinFields {
			refData := doc[field.Code]
			if refData == nil {
				continue
			}
			if field.IsArray {
				// 数组
				if v, ok := refData.([]string); ok {
					arr := []interface{}{}
					for _, str := range v {
						if str == "" {
							continue
						}
						arr = append(arr, bson.ObjectIdHex(str))
					}
					doc[field.Code] = arr
				} else if v, ok := refData.([]bson.ObjectId); ok {
					doc[field.Code] = v
				} else {
					RefModel := kuu.Model(field.JoinName)
					var refDocs []interface{}
					kuu.JSONConvert(doc[field.Code], &refDocs)
					arr := []bson.ObjectId{}
					newDocs := []kuu.H{}
					for _, refItem := range refDocs {
						var (
							refDoc   kuu.H
							existsID bson.ObjectId
						)
						kuu.JSONConvert(refItem, &refDoc)
						if v, ok := refDoc["_id"].(string); ok && v != "" {
							existsID = bson.ObjectIdHex(v)
							arr = append(arr, existsID)
						} else if v, ok := refDoc["_id"].(bson.ObjectId); ok && v != "" {
							existsID = v
							arr = append(arr, existsID)
						} else {
							id := bson.NewObjectId()
							refDoc["_id"] = id
							arr = append(arr, id)
							newDocs = append(newDocs, refDoc)
						}
						if existsID != "" {
							// 作更新操作
							var existsDoc kuu.H
							kuu.JSONConvert(refDoc, &existsDoc)
							delete(existsDoc, "_id")
							if len(existsDoc) > 0 {
								if err := RefModel.Update(kuu.H{"_id": existsID}, refDoc); err != nil {
									kuu.Error(err)
								}
							}
						}
					}
					if len(newDocs) > 0 && (field.Code != "CreatedBy" && field.Code != "UpdatedBy") {
						if _, err := RefModel.Create(newDocs); err != nil {
							kuu.Error(err)
						}
					}
					doc[field.Code] = arr
				}
			} else {
				// 单个
				if v, ok := refData.(string); ok && v != "" {
					doc[field.Code] = bson.ObjectIdHex(v)
				} else if v, ok := refData.(bson.ObjectId); ok && v != "" {
					doc[field.Code] = v
				} else {
					var refDoc kuu.H
					kuu.JSONConvert(doc[field.Code], &refDoc)
					if refDoc == nil {
						continue
					}
					if v, ok := refDoc["_id"].(string); ok && v != "" {
						doc[field.Code] = bson.ObjectIdHex(v)
					} else if v, ok := refDoc["_id"].(bson.ObjectId); ok && v != "" {
						doc[field.Code] = v
					} else if field.Code != "CreatedBy" && field.Code != "UpdatedBy" {
						refDoc["_id"] = bson.NewObjectId()
						doc[field.Code] = refDoc["_id"]
						RefModel := kuu.Model(field.JoinName)
						if _, err := RefModel.Create(refDoc); err != nil {
							kuu.Error(err)
						}
					}
				}
			}
		}
		docs[index] = doc
	}
	return docs
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
	}
}

// 查询引用数据
func findJoinData(session *mgo.Session, field kuu.SchemaField, selector interface{}, result *[]kuu.H) {
	if session == nil || field.Code == "" {
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
