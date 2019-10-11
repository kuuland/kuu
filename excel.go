package kuu

import (
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/jinzhu/gorm"
	"net/url"
	"reflect"
	"strings"
)

const (
	XLSXExport = "XLSX_EXPORT"
	XLSXImport = "XLSX_IMPORT"
)

// ExcelTemplate
type ExcelTemplate struct {
	Model        `rest:"*" displayName:"Excel模板"`
	Type         string                `name:"模板类型" gorm:"not null"`
	Code         string                `name:"模板编码" gorm:"not null"`
	Name         string                `name:"模板名称" gorm:"not null"`
	LocaleKey    string                `name:"国际化键" gorm:"not null"`
	Headers      []ExcelTemplateHeader `gorm:"foreignkey:TemplateID"`
	codeIndexMap map[string]int
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
	LocaleKey  string `name:"国际化键" gorm:"not null"`
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
		failedMessage = c.L("rest_export_failed", "Export failed")
		value         = reflect.New(reflectType).Interface()
		meta          = Meta(value)
		template      ExcelTemplate
		sheetName     = "Sheet1"
	)

	template, err := GetExcelTemplate(reflectType, XLSXExport)
	if err != nil {
		c.STDErr(failedMessage, err)
		return
	}
	file := excelize.NewFile()
	// 初始化列头
	for _, header := range template.Headers {
		val := c.L(header.LocaleKey, header.Header).Render()
		err := file.SetCellValue(sheetName, fmt.Sprintf("%s1", header.xaxis), val)
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}
	}
	// 初始化行数据
	if indirectValue := indirectValue(ret.List); indirectValue.Kind() == reflect.Slice {
		for i := 0; i < indirectValue.Len(); i++ {
			var (
				indirectElem      = indirectValue.Index(i)
				validFormatMethod = false
				row               []string
				mapValues         map[string]interface{}
			)
			if indirectElem.CanAddr() && indirectElem.Kind() != reflect.Ptr {
				indirectElem = indirectElem.Addr()
			}
			if methodValue := indirectElem.MethodByName("ExcelRowFormat"); methodValue.IsValid() {
				unsupportedErr := errors.New("unsupported function ExcelRowFormat")
				switch method := methodValue.Interface().(type) {
				case func(ExcelTemplate) []string:
					if row = method(template); len(row) == 0 {
						err = unsupportedErr
					} else {
						validFormatMethod = true
					}
				case func(ExcelTemplate) map[string]interface{}:
					if mapValues = method(template); mapValues == nil || len(mapValues) == 0 {
						err = unsupportedErr
					} else {
						validFormatMethod = true
					}
				default:
					err = unsupportedErr
					return
				}
				if err != nil {
					c.STDErr(failedMessage, err)
					return
				}
			}
			var scope *gorm.Scope
			if !validFormatMethod {
				scope = DB().NewScope(indirectElem.Interface())
			}
			for index, header := range template.Headers {
				var value interface{}
				if validFormatMethod {
					if len(row) > 0 {
						if index < len(row) {
							value = row[index]
						}
					} else if len(mapValues) > 0 {
						value = mapValues[header.Key]
					}
				} else {
					if field, ok := scope.FieldByName(header.Key); ok {
						value = field.Field.Interface()
					}
				}
				if !IsBlank(value) {
					err := file.SetCellValue(sheetName, fmt.Sprintf("%s%d", header.xaxis, i+2), value)
					if err != nil {
						c.STDErr(failedMessage, err)
						return
					}
				}
			}
		}
	}
	displayName := c.L(template.LocaleKey, meta.DisplayName).Render()
	fileName := fmt.Sprintf("%s.xlsx", displayName)
	fileName = url.QueryEscape(fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")
	_ = file.Write(c.Writer)
}

func GetExcelTemplate(reflectType reflect.Type, actionType string) (template ExcelTemplate, err error) {
	var (
		value = reflect.New(reflectType).Interface()
		meta  = Meta(value)
	)
	// 优先查询数据库配置
	db := DB().First(&template, "code = ? and type = ?", meta.Name, actionType)
	if !db.RecordNotFound() && db.Error != nil {
		err = db.Error
		return
	}
	// 再查询代码配置
	if len(template.Headers) == 0 {
		reflectValue := indirectValue(value)
		if reflectValue.CanAddr() && reflectValue.Kind() != reflect.Ptr {
			reflectValue = reflectValue.Addr()
		}

		if methodValue := reflectValue.MethodByName("ExcelTemplate"); methodValue.IsValid() {
			switch method := methodValue.Interface().(type) {
			case func(string) ExcelTemplate:
				template = method(actionType)
			case func() ExcelTemplate:
				template = method()
			default:
				err = errors.New("unsupported function ExcelTemplate")
				return
			}
		}
	}
	// 最后生成默认配置
	if len(template.Headers) == 0 {
		names := GetXAxisNames(len(meta.Fields))
		var headers []ExcelTemplateHeader
		for index, field := range meta.Fields {
			headers = append(headers, ExcelTemplateHeader{
				Header:    field.Name,
				LocaleKey: field.LocaleKey,
				Key:       field.Code,
				xaxis:     names[index],
			})
		}
		template = ExcelTemplate{
			Headers:   headers,
			LocaleKey: meta.LocaleKey,
		}
	}
	for index, header := range template.Headers {
		if template.codeIndexMap == nil {
			template.codeIndexMap = make(map[string]int)
		}
		if header.LocaleKey == "" {
			header.LocaleKey = strings.ToLower(fmt.Sprintf("kuu_%s_%s", meta.Name, header.Key))
		}
		template.codeIndexMap[header.Key] = index
		template.Headers[index] = header
	}
	if template.LocaleKey == "" {
		template.LocaleKey = strings.ToLower(fmt.Sprintf("menu_%s_doc", meta.Name))
	}
	template.Type = actionType
	return
}

// ExcelImport
func ExcelImport(c *Context, reflectType reflect.Type) {
	failedMessage := c.L("rest_import_failed", "Import failed")
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
	// 解析Excel
	f, err := excelize.OpenReader(src)
	if err != nil {
		c.STDErr(failedMessage, err)
		return
	}
	// 查询模板
	template, err := GetExcelTemplate(reflectType, XLSXImport)
	if err != nil {
		c.STDErr(failedMessage, err)
		return
	}
	activeSheetName := f.GetSheetName(f.GetActiveSheetIndex())
	rows, err := f.GetRows(activeSheetName)
	if err != nil {
		c.STDErr(failedMessage, err)
		return
	}
	var docs []interface{}
	err = c.WithTransaction(func(tx *gorm.DB) error {
		parseErr := errors.New("parse row data failed")
		for index, row := range rows {
			if index == 0 {
				continue
			}
			doc := reflect.New(reflectType).Interface()
			reflectValue := indirectValue(doc)
			if reflectValue.CanAddr() && reflectValue.Kind() != reflect.Ptr {
				reflectValue = reflectValue.Addr()
			}
			if methodValue := reflectValue.MethodByName("ExcelRowParse"); methodValue.IsValid() {
				switch method := methodValue.Interface().(type) {
				case func([]string, ExcelTemplate) *LanguageMessage:
					if msg := method(row, template); msg != nil {
						failedMessage = msg
						return parseErr
					}
				case func([]string, ExcelTemplate) error:
					if err := method(row, template); err != nil {
						return err
					}
				case func([]string, ExcelTemplate) (*LanguageMessage, error):
					if msg, err := method(row, template); msg != nil {
						failedMessage = msg
						return err
					}
				default:
					return errors.New("unsupported function ExcelRowParse")
				}
			} else {
				return errors.New("undefined ExcelRowParse")
			}

			if err := tx.Create(doc).Error; err != nil {
				return err
			}
			docs = append(docs, doc)
		}
		return tx.Error
	})
	if err != nil {
		c.STDErr(failedMessage, err)
		return
	}
	c.STD(docs)
}
