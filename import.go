package kuu

import (
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	uuid "github.com/satori/go.uuid"
	"strconv"
)

var importCallback = make(map[string]*ImportCallbackProcessor)

// ImportCallbackResult
type ImportCallbackResult struct {
	Feedback [][]string
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

// ImportRecord
type ImportRecord struct {
	Model    `rest:"*" displayName:"导入记录"`
	ImportSn string `name:"批次编号" sql:"index"`
	Context  string `name:"导入时上下文数据" gorm:"type:text"`
	Channel  string `name:"导入渠道"`
	Data     string `name:"导入数据（[][]string）" gorm:"type:text"`
	Feedback string `name:"反馈记录（[][]string）" gorm:"type:text"`
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

// ImportCallbackProcessor
type ImportCallbackProcessor func(context *ImportContext, rows [][]string) *ImportCallbackResult

// RegisterImportCallback 注册导入回调
func RegisterImportCallback(channel string, processor ImportCallbackProcessor) {
	importCallback[channel] = &processor
}

// ReimportRecord
func ReimportRecord(importSn string) error {
	var record ImportRecord
	if err := DB().Where(&ImportRecord{ImportSn: importSn}).First(&record).Error; err != nil {
		return err
	}
	if record.Status == ImportStatusImporting {
		return fmt.Errorf("record are being imported: %s", importSn)
	}
	if importCallback[record.Channel] == nil {
		return fmt.Errorf("no import callback registered for this channel: %s", record.Channel)
	}
	if record.Status == ImportStatusImporting {
		return fmt.Errorf("record are being imported: %s", importSn)
	}
	db := DB().Model(&ImportRecord{}).Where(&ImportRecord{Model: Model{ID: record.ID}})
	if err := db.Update(&ImportRecord{Status: ImportStatusImporting}).Error; err != nil {
		return err
	}
	go CallImportCallback(&record)
	return nil
}

// CallImportCallback
func CallImportCallback(info *ImportRecord) {
	if info == nil {
		return
	}

	var rows [][]string
	if err := Parse(info.Data, &rows); err != nil {
		ERROR(err)
		return
	}

	callback := importCallback[info.Channel]
	if callback == nil {
		ERROR("no import callback registered for this channel: %s", info.Channel)
		return
	}

	context := &ImportContext{}
	_ = Parse(info.Context, context)
	result := (*callback)(context, rows)
	if result == nil {
		ERROR("import callback result is nil: %s", info.ImportSn)
		return
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
		doc.Feedback = Stringify(result.Feedback)
	}
	if len(result.Extra) > 0 {
		doc.Extra = Stringify(result.Extra)
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
		if importCallback[channel] == nil {
			c.STDErr(failedMessage, fmt.Errorf("no import callback registered for this channel: %s", channel))
			return
		}
		src, err := file.Open()
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		defer ERROR(src.Close())
		// 解析Excel
		f, err := excelize.OpenReader(src)
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		// 选择工作表
		var sheetName string
		if v := c.PostForm("sheet_name"); v != "" {
			sheetName = v
		}
		if v := c.PostForm("sheet_idx"); sheetName == "" && v != "" {
			idx, err := strconv.Atoi(v)
			if err == nil {
				sheetName = f.GetSheetName(idx)
			}
		}
		if sheetName == "" {
			sheetName = f.GetSheetName(f.GetActiveSheetIndex())
		}
		if sheetName == "" {
			sheetName = f.GetSheetName(1)
		}
		// 读取当前工作表所有行
		rows, err := f.GetRows(sheetName)
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		if len(rows) == 0 {
			c.STDErr(c.L("import_empty", "Import data is empty"))
			return
		}
		// 创建导入记录
		record := ImportRecord{
			Channel: channel,
			Data:    Stringify(rows),
			Context: Stringify(&ImportContext{
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
		go CallImportCallback(&record)
		// 响应请求
		c.STD("ok")
	},
}

// ReimportRoute
var ReimportRoute = RouteInfo{
	Name:   "重新导入路由",
	Method: "GET",
	Path:   "/reimport",
	HandlerFunc: func(c *Context) {
		var (
			failedMessage = c.L("reimport_failed", "Reimport failed")
			importSn      = c.Query("import_sn")
		)
		if importSn == "" {
			c.STDErr(failedMessage, errors.New("'import_sn' is required"))
			return
		}

		if err := ReimportRecord(importSn); err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		c.STD("ok")
	},
}
