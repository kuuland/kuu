package kuu

import (
	"github.com/kuuland/kuu/intl"
	"net/http"
	"strconv"
	"strings"
)

var IntlLanguagesRoute = RouteInfo{
	Name:   "查询语言列表",
	Method: http.MethodGet,
	Path:   "/intl/languages",
	HandlerFunc: func(c *Context) {
		list := intl.LanguageList()
		c.STD(list)
	},
}

var IntlMessagesRoute = RouteInfo{
	Name:   "查询消息列表",
	Method: http.MethodGet,
	Path:   "/intl/messages",
	HandlerFunc: func(c *Context) {
		var query struct {
			LanguageCodes string `form:"langs"`
			Prefix        string `form:"prefix"`
			Description   string `form:"desc"`
			Keys          string `form:"keys"`
		}
		_ = c.ShouldBindQuery(&query)
		messagesMap := getIntlMessages(&IntlMessagesOptions{
			LanguageCodes: query.LanguageCodes,
			Prefix:        query.Prefix,
			Description:   query.Description,
			Keys:          query.Keys,
		})
		c.STD(messagesMap)
	},
}

var IntlMessagesSaveRoute = RouteInfo{
	Name:   "修改/新增翻译键",
	Method: http.MethodPost,
	Path:   "/intl/messages/save",
	HandlerFunc: func(c *Context) {
		failedMessage := c.L("intl_messages_save_failed", "Save failed.")
		var messages map[string]map[string]string
		if err := c.ShouldBindJSON(&messages); err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		if err := saveIntlMessages(messages, false); err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		c.STDOK()
	},
}

var IntlMessagesUploadRoute = RouteInfo{
	Name:   "批量上传翻译文件",
	Method: http.MethodPost,
	Path:   "/intl/messages/upload",
	HandlerFunc: func(c *Context) {
		failedMessage := c.L("intl_messages_upload_failed", "Upload failed.")

		updateMethod := c.DefaultPostForm("method", "incr")
		fh, err := c.FormFile("file")
		if err != nil {
			c.STDErr(failedMessage, err)
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
		rows, err := ParseExcelFromFileHeader(fh, sheetIndex, sheetName)
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		if len(rows) == 0 {
			c.STDOK()
			return
		}
		languages := intl.LanguageList()
		indexLangCodeMap := map[int]string{
			0: "key",
			1: "default",
			2: "en",
			3: "zh-Hans",
			4: "zh-Hant",
		}
		i := len(indexLangCodeMap)
		for _, item := range languages {
			if item.Code == "en" || item.Code == "zh-Hans" || item.Code == "zh-Hant" {
				continue
			}
			indexLangCodeMap[i] = item.Code
			i++
		}
		messages := make(map[string]map[string]string)
		for i := 1; i < len(rows); i++ {
			row := rows[i]
			key := strings.TrimSpace(row[0])
			if key == "" {
				continue
			}
			for j := 1; j < len(row); j++ {
				value := strings.TrimSpace(row[j])
				if value == "" {
					continue
				}
				lang := indexLangCodeMap[j]
				if messages[key] == nil {
					messages[key] = make(map[string]string)
				}
				messages[key][lang] = value
			}
		}
		if err := saveIntlMessages(messages, updateMethod == "full"); err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		c.STDOK()
	},
}
