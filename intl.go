package kuu

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/kuuland/kuu/intl"
	"io/ioutil"
	"net/http"
	"os"
	"path"
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
			Keys          string `form:"keys"`
		}
		query.LanguageCodes = strings.ToLower(query.LanguageCodes)
		query.Prefix = strings.ToLower(query.Prefix)
		query.Keys = strings.ToLower(query.Keys)
		_ = c.ShouldBindQuery(&query)
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
						if messagesMap[keyStr] == nil {
							messagesMap[keyStr] = make(map[string]string)
						}
						if query.Prefix != "" && !strings.Contains(lowerCaseKeyStr, query.Prefix) {
							return nil
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
				// 更新内容
				for k, values := range messages {
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
			c.STDErr(failedMessage, errs)
			return
		}
		c.STDOK()
	},
}

var IntlMessagesUpload = RouteInfo{
	Name:   "上传单语言翻译文件（完全覆盖）",
	Method: http.MethodPost,
	Path:   "/intl/messages/upload",
	HandlerFunc: func(c *Context) {
		failedMessage := c.L("intl_messages_upload_failed", "Upload failed.")
		langCode := c.PostForm("lang")
		fh, err := c.FormFile("file")
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}

		r, err := fh.Open()
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		defer func() { _ = r.Close() }()
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		if err := saveMessageFile(langCode, buf); err != nil {
			c.STDErr(failedMessage, err)
			return
		}
		c.STDOK()
	},
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
