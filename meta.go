package kuu

import (
	"path"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	metadataMap     = make(map[string]*Metadata)
	metadataList    = make([]*Metadata, 0)
	modelStructsMap sync.Map
)

// Metadata
type Metadata struct {
	ModCode       string
	Name          string
	DisplayName   string
	FullName      string
	Fields        []MetadataField
	RestDesc      *RestDesc    `json:"-"`
	reflectType   reflect.Type `json:"-"`
	SubDocIDNames []string     `json:"-" gorm:"-"`
	UIDNames      []string     `json:"-" gorm:"-"`
	OrgIDNames    []string     `json:"-" gorm:"-"`
}

// MetadataField
type MetadataField struct {
	Code       string
	Name       string
	Kind       string
	Type       string
	Value      interface{} `json:"-" gorm:"-"`
	Enum       string
	IsRef      bool
	IsPassword bool
	IsArray    bool
}

// NewValue
func (m *Metadata) NewValue() interface{} {
	return reflect.New(m.reflectType).Interface()
}

// OmitPassword
func (m *Metadata) OmitPassword(data interface{}) interface{} {
	if m == nil {
		return data
	}

	var passwordKeys []string
	for _, field := range m.Fields {
		if field.IsPassword {
			passwordKeys = append(passwordKeys, field.Code)
		}
	}
	if len(passwordKeys) == 0 {
		return data
	}

	execOmit := func(indirectValue reflect.Value) {
		var val interface{}
		if indirectValue.CanAddr() {
			val = indirectValue.Addr().Interface()
		} else {
			val = indirectValue.Interface()
		}
		scope := DB().NewScope(val)
		for _, key := range passwordKeys {
			if _, ok := scope.FieldByName(key); ok {
				if err := scope.SetColumn(key, ""); err != nil {
					ERROR(err)
				}
			}
		}
	}
	if indirectValue := indirect(reflect.ValueOf(data)); indirectValue.Kind() == reflect.Slice {
		for i := 0; i < indirectValue.Len(); i++ {
			execOmit(indirectValue.Index(i))
		}
	} else {
		execOmit(indirectValue)
	}
	return data
}

func parseMetadata(value interface{}) (m *Metadata) {
	reflectType := reflect.ValueOf(value).Type()
	for reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}

	// Scope value need to be a struct
	if reflectType.Kind() != reflect.Struct {
		return
	}

	hashKey := reflectType
	if value, ok := modelStructsMap.Load(hashKey); ok && value != nil {
		return value.(*Metadata)
	}

	reflectTypeName := reflectType.Name()
	m = &Metadata{
		Name:        reflectTypeName,
		FullName:    path.Join(reflectType.PkgPath(), reflectTypeName),
		reflectType: reflectType,
	}
	for i := 0; i < reflectType.NumField(); i++ {
		fieldStruct := reflectType.Field(i)
		displayName := fieldStruct.Tag.Get("displayName")
		if m.DisplayName == "" && displayName != "" {
			m.DisplayName = displayName
		}
		indirectType := fieldStruct.Type
		for indirectType.Kind() == reflect.Ptr {
			indirectType = indirectType.Elem()
		}
		fieldValue := reflect.New(indirectType).Interface()
		field := MetadataField{
			Code: fieldStruct.Name,
			Kind: fieldStruct.Type.Kind().String(),
			Enum: fieldStruct.Tag.Get("enum"),
		}
		switch field.Kind {
		case "bool":
			field.Type = "boolean"
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64":
			field.Type = "integer"
		case "float32", "float64":
			field.Type = "number"
		case "slice", "struct", "ptr":
			field.Type = "object"
		default:
			field.Type = field.Kind
		}
		if _, ok := fieldValue.(*time.Time); ok {
			field.Type = "string"
		}
		ref := fieldStruct.Tag.Get("ref")
		if ref != "" {
			fieldMeta := Meta(ref)
			if fieldMeta != nil {
				field.Type = fieldMeta.Name
				field.IsRef = true
				field.Value = fieldValue
				if indirectType.Kind() == reflect.Slice {
					field.IsArray = true
				}
			}
		}
		tagSettings := parseTagSetting(fieldStruct.Tag, "kuu")
		if len(tagSettings) > 0 {
			if _, exists := tagSettings["PASSWORD"]; exists {
				field.IsPassword = true
			}
			if v, exists := tagSettings["UIDS"]; exists {
				m.UIDNames = strings.Split(v, ",")
			}
			if v, exists := tagSettings["SUB_IDS"]; exists {
				m.SubDocIDNames = strings.Split(v, ",")
			}
			if v, exists := tagSettings["ORG_IDS"]; exists {
				m.OrgIDNames = strings.Split(v, ",")
			}
		}

		name := fieldStruct.Tag.Get("name")
		if name != "" {
			field.Name = name
		}
		if field.Name != "" {
			m.Fields = append(m.Fields, field)
		}
	}
	modelStructsMap.Store(hashKey, m)
	metadataMap[m.Name] = m
	metadataList = append(metadataList, m)
	return
}

// Meta
func Meta(valueOrName interface{}) (m *Metadata) {
	if v, ok := valueOrName.(string); ok {
		return metadataMap[v]
	} else {
		return parseMetadata(valueOrName)
	}
}

// Metalist
func Metalist() []*Metadata {
	return metadataList
}

// RegisterMeta
func RegisterMeta() {
	tx := DB().Begin()
	tx = tx.Unscoped().Where(&Metadata{}).Delete(&Metadata{})
	for _, meta := range metadataList {
		tx = tx.Create(meta)
	}
	if errs := tx.GetErrors(); len(errs) > 0 {
		ERROR(errs)
		if err := tx.Rollback(); err != nil {
			ERROR(err)
		}
	} else {
		if err := tx.Commit().Error; err != nil {
			ERROR(err)
		}
	}
}
