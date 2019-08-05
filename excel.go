package kuu

import (
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/jinzhu/gorm"
	"net/url"
	"reflect"
)

const (
	XLSXExport = "XLSX_EXPORT"
	XLSXImport = "XLSX_IMPORT"
)

// ExcelTemplate
type ExcelTemplate struct {
	Model   `rest:"*" displayName:"Excel模板"`
	Code    string                `name:"模板编码" gorm:"not null"`
	Name    string                `name:"模板名称" gorm:"not null"`
	Headers []ExcelTemplateHeader `gorm:"foreignkey:TemplateID"`
}

// BeforeFind
func (t *ExcelTemplate) BeforeFind(scope *gorm.Scope) {
	scope.Search.Preload("Headers")
}

// BeforeSave
func (t *ExcelTemplate) BeforeSave() {
	if t.Code == "" {
		t.Code = RandCode()
	}
}

// ExcelTemplateHeader
type ExcelTemplateHeader struct {
	Model      `rest:"*" displayName:"Excel模板列"`
	TemplateID uint
	Header     string `name:"列标题" gorm:"not null"`
	Key        string `name:"字段名" gorm:"not null"`
	xaxis      string
}

func GetXAxisNames(size int) []string {
	var names []string
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i, out := range fmt.Sprintf(" %s", letters) {
		first := string(out)
		for j, in := range letters {
			second := string(in)
			if count := i*len(letters) + j; count >= size {
				break
			}
			name := If(first == " ", string(second), fmt.Sprintf("%s%s", string(first), string(second))).(string)
			names = append(names, name)
		}
	}
	return names
}

// ExcelExport
func ExcelExport(c *Context, ret *BizQueryResult, reflectType reflect.Type) {
	var (
		failedMessage = L("rest_export_failed", "Export failed")
		value         = reflect.New(reflectType).Interface()
		meta          = Meta(value)
		template      ExcelTemplate
		sheetName     = "Sheet1"
		codeIndexMap  = make(map[string]int)
	)

	// 优先查询数据库配置
	db := c.DB().First(&template, "code = ?", fmt.Sprintf("%sExport", meta.Name))
	if err := db.Error; !db.RecordNotFound() && err != nil {
		c.STDErr(failedMessage, err)
		return
	}
	// 再查询代码配置
	if db.NewRecord(template) {
		reflectValue := indirectValue(value)
		if reflectValue.CanAddr() && reflectValue.Kind() != reflect.Ptr {
			reflectValue = reflectValue.Addr()
		}

		if methodValue := reflectValue.MethodByName("ExcelTemplate"); methodValue.IsValid() {
			switch method := methodValue.Interface().(type) {
			case func(string) ExcelTemplate:
				template = method(XLSXExport)
			default:
				c.STDErr(failedMessage, errors.New("unsupported function ExcelTemplate"))
			}
		}
	}
	// 最后生成默认配置
	if db.NewRecord(template) {
		names := GetXAxisNames(len(meta.Fields))
		var headers []ExcelTemplateHeader
		for index, field := range meta.Fields {
			headers = append(headers, ExcelTemplateHeader{
				Header: c.L(If(field.NameLocaleKey != "", field.NameLocaleKey, field.Name).(string), field.Name).Render(),
				Key:    field.Code,
				xaxis:  names[index],
			})
			codeIndexMap[field.Code] = index
		}
		template = ExcelTemplate{Headers: headers}
	}
	file := excelize.NewFile()
	// 初始化列头
	for _, header := range template.Headers {
		err := file.SetCellValue(sheetName, fmt.Sprintf("%s1", header.xaxis), header.Header)
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}
	}
	// 初始化行数据
	if indirectValue := indirectValue(ret.List); indirectValue.Kind() == reflect.Slice {
		for i := 0; i < indirectValue.Len(); i++ {
			indirectValue := indirectValue.Index(i)
			var val interface{}
			if indirectValue.CanAddr() {
				val = indirectValue.Addr().Interface()
			} else {
				val = indirectValue.Interface()
			}
			scope := DB().NewScope(val)
			for _, header := range template.Headers {
				if field, ok := scope.FieldByName(header.Key); ok {
					value := field.Field.Interface()
					err := file.SetCellValue(sheetName, fmt.Sprintf("%s%d", header.xaxis, i+2), value)
					if err != nil {
						c.STDErr(failedMessage, err)
						return
					}
				}

			}
		}
	}
	displayName := c.L(If(meta.DisplayNameLocaleKey != "", meta.DisplayNameLocaleKey, meta.DisplayName).(string), meta.DisplayName).Render()
	fileName := fmt.Sprintf("%s.xlsx", displayName)
	fileName = url.QueryEscape(fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")
	_ = file.Write(c.Writer)
}

// ExcelImport
func ExcelImport(c *Context, reflectType reflect.Type) {
	var (
		failedMessage = L("rest_import_failed", "Import failed")
	)
	// 解析请求体
	file, _ := c.FormFile("file")
	if file == nil {
		c.STDErr(failedMessage, errors.New("no 'file' key in form-data"))
		return
	}
	src, err := file.Open()
	if err != nil {
		c.STDErr(failedMessage, err)
		return
	}
	defer src.Close()
	// TODO 解析Excel

	// TODO 插入数据库

}
