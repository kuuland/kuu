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
	NativeName    string
	DisplayName   string
	LocaleKey     string
	FullName      string
	Fields        []MetadataField
	RestDesc      *RestDesc `json:"-"`
	reflectType   reflect.Type
	SubDocIDNames []string          `json:"-" gorm:"-"`
	UIDNames      []string          `json:"-" gorm:"-"`
	OrgIDNames    []string          `json:"-" gorm:"-"`
	TagSettings   map[string]string `json:"-" gorm:"-"`
}

// MetadataField
type MetadataField struct {
	Code         string
	Name         string
	NativeName   string
	DBType       string
	IsBland      bool
	IsPrimaryKey bool
	LocaleKey    string
	Kind         string
	Type         string
	Enum         string
	IsRef        bool
	IsPassword   bool
	IsArray      bool
	Value        interface{}       `json:"-" gorm:"-"`
	Tag          reflect.StructTag `json:"-" gorm:"-"`
	TagSetting   map[string]string `json:"-" gorm:"-"`
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
	if indirectValue := indirectValue(data); indirectValue.Kind() == reflect.Slice {
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
		TagSettings: map[string]string{},
	}
	modelScope := DB().NewScope(value)
	m.NativeName = modelScope.TableName()
	for i := 0; i < reflectType.NumField(); i++ {
		fieldStruct := reflectType.Field(i)
		displayName := fieldStruct.Tag.Get("displayName")
		if m.DisplayName == "" && displayName != "" {
			m.DisplayName = displayName
			if v := fieldStruct.Tag.Get("locale"); m.LocaleKey == "" && v != "" {
				m.LocaleKey = v
			}
		}
		indirectType := fieldStruct.Type
		for indirectType.Kind() == reflect.Ptr {
			indirectType = indirectType.Elem()
		}
		fieldValue := reflect.New(indirectType).Interface()
		field := MetadataField{
			Code:      fieldStruct.Name,
			Kind:      fieldStruct.Type.String(),
			Enum:      fieldStruct.Tag.Get("enum"),
			LocaleKey: fieldStruct.Tag.Get("locale"),
			Tag:       fieldStruct.Tag,
		}
		if modelField, hasModelField := modelScope.FieldByName(field.Code); hasModelField {
			field.NativeName = modelField.DBName
			field.IsBland = modelField.IsBlank
			field.IsPrimaryKey = modelField.IsPrimaryKey
			if modelField.IsNormal {
				field.DBType = DB().Dialect().DataTypeOf(modelField.StructField)
			}
		}
		switch field.Kind {
		case "bool", "null.Bool":
			field.Type = "boolean"
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64", "null.Int":
			field.Type = "integer"
		case "float32", "float64":
			field.Type = "number"
		case "slice", "struct", "ptr":
			field.Type = "object"
		case "null.String", "string":
			field.Type = "string"
		case "gorm.Model", "kuu.Model", "kuu.ModelExOrg":
			temp := parseMetadata(fieldValue)
			for _, metadataField := range temp.Fields {
				m.Fields = append(m.Fields, metadataField)
			}

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
			field.TagSetting = tagSettings
			for key, value := range tagSettings {
				m.TagSettings[key] = value
			}
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
		if field.DBType != "" {
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

func parseTagSetting(tags reflect.StructTag, tagKey string) map[string]string {
	setting := map[string]string{}
	str := tags.Get(tagKey)
	split := strings.Split(str, ";")
	for _, value := range split {
		if value == "" {
			continue
		}
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToUpper(v[0]))
		if len(v) >= 2 {
			setting[k] = strings.Join(v[1:], ":")
		} else {
			setting[k] = k
		}
	}
	return setting
}
