package kuu

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/kuuland/kuu/intl"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

func getLocalesDir() string {
	localesDir := "assets/locales"
	if v := C().GetString("localesDir"); v != "" {
		localesDir = v
	}
	return localesDir
}

var IntlLanguages = RouteInfo{
	Name:   "查询语言列表",
	Method: http.MethodGet,
	Path:   "/intl/languages",
	HandlerFunc: func(c *Context) {
		list := intl.LanguageList()
		c.STD(list)
	},
}

var IntlMessages = RouteInfo{
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

var IntlMessagesSave = RouteInfo{
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

var IntlMessagesUpload = RouteInfo{
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

type IntlMessagesOptions struct {
	LanguageCodes string
	Prefix        string
	Description   string
	Keys          string
}

func getIntlMessages(opts ...*IntlMessagesOptions) map[string]map[string]string {
	query := IntlMessagesOptions{}
	if len(opts) > 0 && opts[0] != nil {
		query = *opts[0]
	}
	query.LanguageCodes = strings.ToLower(query.LanguageCodes)
	query.Prefix = strings.ToLower(query.Prefix)
	query.Keys = strings.ToLower(query.Keys)
	localesDir := getLocalesDir()
	fis, err := ioutil.ReadDir(localesDir)
	var messagesMap = make(map[string]map[string]string)
	if err == nil {
		for _, fi := range fis {
			if fi.IsDir() {
				continue
			}
			fileName := fi.Name()
			filePath := path.Join(localesDir, fileName)
			langCode := strings.ReplaceAll(fileName, path.Ext(fileName), "")
			lowerCaseLangCode := strings.ToLower(langCode)
			buf, err := ioutil.ReadFile(filePath)
			if err == nil {
				_ = jsonparser.ObjectEach(buf, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
					keyStr := string(key)
					valueStr := string(value)
					lowerCaseKeyStr := strings.ToLower(keyStr)
					if query.Prefix != "" && !strings.Contains(lowerCaseKeyStr, query.Prefix) {
						return nil
					}
					if messagesMap[keyStr] == nil {
						messagesMap[keyStr] = make(map[string]string)
					}
					if query.Keys != "" && !strings.Contains(query.Keys, lowerCaseKeyStr) {
						return nil
					}
					if query.LanguageCodes != "" && !strings.Contains(query.LanguageCodes, lowerCaseLangCode) {
						return nil
					}
					messagesMap[keyStr][langCode] = valueStr
					return nil
				})
			}
		}
	}
	// 过滤Description
	query.Description = strings.TrimSpace(query.Description)
	if query.Description != "" {
		for k, values := range messagesMap {
			var has bool
			for l, v := range values {
				if l == "default" && strings.Contains(v, query.Description) {
					has = true
				}
			}
			if !has {
				delete(messagesMap, k)
			}
		}
	}
	// 过滤预置键
	for k, presetValues := range presetIntlMessages {
		if k == "" || len(presetValues) == 0 {
			continue
		}
		currentValues := messagesMap[k]
		if len(currentValues) == 0 {
			messagesMap[k] = presetValues
		} else {
			for l, v := range presetValues {
				if vv, has := currentValues[l]; has && vv != "" {
					continue
				}
				messagesMap[k][l] = v
			}
		}
	}

	return messagesMap
}

func saveIntlMessages(messages map[string]map[string]string, replace bool) error {
	if len(messages) == 0 {
		return nil
	}
	var (
		langCodeMap = make(map[string]bool)
		langCodes   []string
	)
	for _, values := range messages {
		for l := range values {
			if langCodeMap[l] {
				continue
			}
			langCodeMap[l] = true
			langCodes = append(langCodes, l)
		}
	}
	if err := ensureLocaleFiles(langCodes); err != nil {
		return err
	}
	languageMap := intl.LanguageMap()
	localesDir := getLocalesDir()
	fis, err := ioutil.ReadDir(localesDir)
	var errs []error
	if err == nil {
		for _, fi := range fis {
			if fi.IsDir() {
				continue
			}
			fileName := fi.Name()
			filePath := path.Join(localesDir, fileName)
			langCode := strings.ReplaceAll(fileName, path.Ext(fileName), "")
			if langCode != "default" && languageMap[langCode] == "" {
				continue
			}
			// 读取原文件
			var currentContent map[string]string
			if replace {
				currentContent = make(map[string]string)
			} else {
				buf, err := ioutil.ReadFile(filePath)
				if os.IsNotExist(err) {
					currentContent = make(map[string]string)
				} else if err != nil {
					errs = append(errs, err)
					continue
				} else {
					if err := JSON().Unmarshal(buf, &currentContent); err != nil {
						errs = append(errs, err)
						continue
					}
				}
			}
			// 更新内容
			for k, values := range messages {
				if _, has := values["_dr"]; has {
					delete(currentContent, k)
					continue
				}
				for l, v := range values {
					if l != langCode {
						continue
					}
					currentContent[k] = v
				}
			}
			// 保存文件
			if err := saveMessageFile(langCode, currentContent); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

func ensureLocaleFiles(langCodes []string) error {
	localesDir := getLocalesDir()
	for _, l := range langCodes {
		if l == "_dr" {
			continue
		}
		filePath := path.Join(localesDir, fmt.Sprintf("%s.json", l))
		if _, err := os.Lstat(filePath); os.IsNotExist(err) {
			if err := ioutil.WriteFile(filePath, []byte("{}"), os.ModePerm); err != nil {
				return err
			}
		}
	}
	return nil
}

func saveMessageFile(langCode string, content interface{}) error {
	languageMap := intl.LanguageMap()
	if langCode != "default" && languageMap[langCode] == "" {
		return fmt.Errorf("incorrect language code: %s", langCode)
	}
	var buf []byte
	switch v := content.(type) {
	case []byte:
		buf = v
	default:
		vv, err := json.Marshal(content)
		if err != nil {
			return err
		}
		buf = vv
	}
	localesDir := getLocalesDir()
	filePath := path.Join(localesDir, fmt.Sprintf("%s.json", langCode))
	if err := ioutil.WriteFile(filePath, buf, os.ModePerm); err != nil {
		return err
	}
	_ = PublishCache("intl_messages_changed", langCode)
	return nil
}
