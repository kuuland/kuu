package kuu

import (
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	uuid "github.com/satori/go.uuid"
	"net/url"
	"strconv"
	"strings"
)

var importCallbackMap = make(map[string]*ImportCallback)

// ImportCallbackResult
type ImportCallbackResult struct {
	Feedback []ImportFeedback
	Message  string
	Error    error
	Extra    map[string]interface{}
}

const (
	ImportStatusImporting = "importing"
	ImportStatusSuccess   = "success"
	ImportStatusFailed    = "failed"
)

func init() {
	Enum("ImportStatus", "导入状态").
		Add(ImportStatusImporting, "Importing").
		Add(ImportStatusSuccess, "Success").
		Add(ImportStatusFailed, "Failed")
}

type ImportContext struct {
	Token      string
	SignType   string
	Lang       string
	UID        uint
	SubDocID   uint
	ActOrgID   uint
	ActOrgCode string
	ActOrgName string
}

type ImportCallback struct {
	Channel           string
	Description       string
	TemplateGenerator func(*Context) (string, []string)
	Validator         ImportCallbackValidator
	Processor         ImportCallbackProcessor
}

type ImportFeedback struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ImportRecord
type ImportRecord struct {
	Model    `rest:"*" displayName:"导入记录"`
	Sync     bool   `name:"是否同步执行"`
	ImportSn string `name:"批次编号" sql:"index"`
	Context  string `name:"导入时上下文数据" gorm:"type:text"`
	Channel  string `name:"导入渠道"`
	Data     string `name:"导入数据（[][]string）" gorm:"type:text"`
	Feedback string `name:"反馈记录（[]string）" gorm:"type:text"`
	Status   string `name:"导入状态" enum:"ImportStatus"`
	Message  string `name:"导入结果"`
	Error    string `name:"错误详情"`
	Extra    string `name:"额外数据（JSON）" gorm:"type:text"`
}

// BeforeCreate
func (imp *ImportRecord) BeforeCreate() {
	imp.ImportSn = uuid.NewV4().String()
	imp.Status = ImportStatusImporting
}

// ImportCallbackValidator
type ImportCallbackValidator func(c *Context, rows [][]string) (*LanguageMessage, error)

// ImportCallbackProcessor
type ImportCallbackProcessor func(c *Context, rows [][]string) *ImportCallbackResult

// RegisterImportCallback 注册导入回调
func RegisterImportCallback(callback *ImportCallback) {
	if callback == nil {
		return
	}

	importCallbackMap[callback.Channel] = callback
}

// CallImportCallback
func CallImportCallback(c *Context, info *ImportRecord) {
	if info == nil {
		return
	}

	var rows [][]string
	if err := JSONParse(info.Data, &rows); err != nil {
		ERROR(err)
		return
	}

	callback := importCallbackMap[info.Channel]
	if callback == nil {
		ERROR("no import callback registered for this channel: %s", info.Channel)
		return
	}
	args := callback
	result := args.Processor(c, rows)
	if result == nil {
		result = &ImportCallbackResult{Message: "success"}
	}
	db := DB().Model(&ImportRecord{}).Where(&ImportRecord{Model: Model{ID: info.ID}})
	doc := ImportRecord{Message: result.Message}
	if result.Error != nil {
		doc.Error = result.Error.Error()
		doc.Status = ImportStatusFailed
	} else {
		doc.Status = ImportStatusSuccess
	}
	if len(result.Feedback) > 0 {
		doc.Feedback = JSONStringify(result.Feedback)
	}
	if len(result.Extra) > 0 {
		doc.Extra = JSONStringify(result.Extra)
	}
	if err := db.Update(&doc).Error; err != nil {
		ERROR(err)
	}
}

// ImportRoute
var ImportRoute = RouteInfo{
	Name:   "统一导入路由",
	Method: "POST",
	Path:   "/import",
	HandlerFunc: func(c *Context) {
		failedMessage := c.L("import_failed", "Import failed")
		// 解析请求体
		file, _ := c.FormFile("file")
		if file == nil {
			c.STDErr(failedMessage, errors.New("no 'file' key in form-data"))
			return
		}
		channel := c.PostForm("channel")
		if channel == "" {
			c.STDErr(failedMessage, errors.New("no 'channel' key in form-data"))
			return
		}
		if importCallbackMap[channel] == nil {
			c.STDErr(failedMessage, fmt.Errorf("no import callback registered for this channel: %s", channel))
			return
		}
		var (
			sheetIndex int
			sheetName  string
		)
		sheetName = c.PostForm("sheet_name")
		if v := c.PostForm("sheet_idx"); v != "" {
			idx, err := strconv.Atoi(v)
			if err == nil {
				sheetIndex = idx
			}
		}
		rows, err := ParseExcelFromFileHeader(file, sheetIndex, sheetName)
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		if len(rows) == 0 {
			c.STDErr(c.L("import_empty", "Import data is empty"))
			return
		}
		// 调用导入验证
		args := importCallbackMap[channel]
		if args.Validator != nil {
			msg, err := args.Validator(c, rows)
			if err != nil {
				if msg == nil {
					msg = failedMessage
				}
				c.STDErr(msg, err)
				return
			}
		}
		// 创建导入记录
		record := ImportRecord{
			Channel: channel,
			Data:    JSONStringify(rows),
			Sync:    c.PostForm("sync") != "",
			Context: JSONStringify(&ImportContext{
				Token:      c.SignInfo.Token,
				SignType:   c.SignInfo.Type,
				Lang:       c.SignInfo.Lang,
				UID:        c.SignInfo.UID,
				SubDocID:   c.SignInfo.SubDocID,
				ActOrgID:   c.PrisDesc.ActOrgID,
				ActOrgCode: c.PrisDesc.ActOrgCode,
				ActOrgName: c.PrisDesc.ActOrgName,
			}),
		}
		if err := c.DB().Create(&record).Error; err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		// 触发导入回调
		if record.Sync {
			CallImportCallback(c, &record)
		} else {
			go CallImportCallback(c, &record)
		}
		// 响应请求
		c.STD(record.ImportSn)
	},
}

// ImportTemplateRoute
var ImportTemplateRoute = RouteInfo{
	Name:   "导入模板下载",
	Method: "GET",
	Path:   "/import/template",
	HandlerFunc: func(c *Context) {
		failedMessage := c.L("import_template_failed", "Import template download failed")
		channel := c.Query("channel")
		format := strings.ToLower(c.DefaultQuery("format", "file"))
		if channel == "" {
			c.STDErr(failedMessage, errors.New("no 'channel' key in query parameters"))
			return
		}
		callback := importCallbackMap[channel]
		if callback == nil {
			c.STDErr(failedMessage, fmt.Errorf("no import callback registered for this channel: %s", channel))
			return
		}
		if callback.TemplateGenerator == nil {
			c.STDErr(failedMessage, fmt.Errorf("no template generator registered for this channel: %s", channel))
			return
		}
		fileName, headers := callback.TemplateGenerator(c)
		switch format {
		case "json":
			c.STD(headers)
		case "file":
			if !strings.HasSuffix(fileName, ".xlsx") {
				fileName = fmt.Sprintf("%s.xlsx", fileName)
			}
			fileName = url.QueryEscape(fileName)
			f := excelize.NewFile()
			if err := f.SetSheetRow("Sheet1", "A1", &headers); err != nil {
				c.STDErr(failedMessage, err)
				return
			}
			c.Header("Content-Transfer-Encoding", "binary")
			c.Header("Content-Disposition", "attachment; filename="+fileName)
			c.Header("Content-Type", "application/octet-stream")
			f.SetActiveSheet(1)
			if err := f.Write(c.Writer); err != nil {
				c.STDErr(failedMessage, err)
				return
			}
		}
	},
}
