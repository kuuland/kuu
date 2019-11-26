package kuu

import (
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"time"
)

// ImportCallbackArgs
type ImportCallbackArgs struct {
	Channel        string
	FirstSheetRows [][]string
	File           *excelize.File
	Context        *Context
}

// ImportCallbackResult
type ImportCallbackResult struct {
	ImportData   [][]string
	FeedbackData [][]string
	Message      string
	Error        error
	Extra        map[string]interface{}
}

// ImportLog
type ImportLog struct {
	Model    `rest:"*" displayName:"系统导入日志"`
	Channel  string `name:"导入渠道"`
	Data     string `name:"导入数据（[][]string）" gorm:"type:text"`
	Feedback string `name:"反馈记录（[][]string）" gorm:"type:text"`
	Message  string `name:"导入结果"`
	Error    string `name:"错误详情"`
	Extra    string `name:"额外数据（JSON）" gorm:"type:text"`
}

// ImportCallbackProcessor
type ImportCallbackProcessor func(*ImportCallbackArgs) *ImportCallbackResult

// ImportCallbacks
var ImportCallbacks = make(map[string][]*ImportCallbackProcessor)

// RegisterImportCallback 注册导入回调
func RegisterImportCallback(channel string, processor ImportCallbackProcessor) {
	ImportCallbacks[channel] = append(ImportCallbacks[channel], &processor)
}

func callImportCallback(args *ImportCallbackArgs) {
	if args == nil {
		return
	}

	callbacks := ImportCallbacks[args.Channel]
	for _, fn := range callbacks {
		result := (*fn)(args)
		if result == nil {
			continue
		}
		info := &ImportLog{
			Channel: args.Channel,
			Message: result.Message,
		}
		if result.Error != nil {
			if info.Message == "" {
				info.Message = args.Context.L("import_failed", "Import failed").Render()
			}
			info.Error = result.Error.Error()
		} else {
			if info.Message == "" {
				info.Message = args.Context.L("import_success", "Imported successfully").Render()
			}
		}
		if len(result.ImportData) > 0 {
			info.Data = Stringify(result.ImportData)
		} else if len(args.FirstSheetRows) > 0 {
			info.Data = Stringify(args.FirstSheetRows)
		}
		if len(result.FeedbackData) > 0 {
			info.Feedback = Stringify(result.FeedbackData)
		}
		if len(result.Extra) > 0 {
			info.Extra = Stringify(result.Extra)
		}
		info.Model.CreatedAt = time.Now()
		info.Model.UpdatedAt = time.Now()
		if args.Context != nil {
			signInfo := args.Context.SignInfo
			prisDesc := args.Context.PrisDesc
			if signInfo != nil {
				if info.Model.CreatedByID == 0 {
					info.Model.CreatedByID = signInfo.UID
				}
				if info.Model.UpdatedByID == 0 {
					info.Model.UpdatedByID = signInfo.UID
				}
			}
			if prisDesc != nil {
				if info.Model.OrgID == 0 {
					info.Model.OrgID = prisDesc.ActOrgID
				}
			}
		}
		if err := DB().Create(&info).Error; err != nil {
			ERROR(err)
		}
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
		if len(ImportCallbacks[channel]) == 0 {
			c.STDErr(failedMessage, fmt.Errorf("no callback registered for this channel: %s", channel))
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
		activeSheetName := f.GetSheetName(f.GetActiveSheetIndex())
		if activeSheetName == "" {
			activeSheetName = f.GetSheetName(1)
		}
		rows, err := f.GetRows(activeSheetName)
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		go callImportCallback(&ImportCallbackArgs{
			Channel:        channel,
			FirstSheetRows: rows,
			File:           f,
			Context:        c,
		})
		c.STD("ok")
	},
}
