package kuu

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/kuuland/kuu/intl"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
)

const intlMessagesChangedChannel = "intl_messages_changed"

var (
	intlMessages   = make(map[string]map[string]string)
	intlMessagesMu sync.RWMutex
)

func GetIntlMessages() map[string]map[string]string {
	intlMessagesMu.RLock()
	defer intlMessagesMu.RUnlock()
	return intlMessages
}

func ReloadIntlMessages() {
	intlMessagesMu.Lock()
	defer intlMessagesMu.Unlock()
	intlMessages = getIntlMessages()
}

func getLocalesDir() string {
	localesDir := "assets/locales"
	if v := C().GetString("localesDir"); v != "" {
		localesDir = v
	}
	return localesDir
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
	query.LanguageCodes = strings.TrimSpace(query.LanguageCodes)
	query.Prefix = strings.TrimSpace(query.Prefix)
	query.Description = strings.TrimSpace(query.Description)
	query.Keys = strings.TrimSpace(query.Keys)

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
			buf, err := ioutil.ReadFile(filePath)
			if err == nil {
				_ = jsonparser.ObjectEach(buf, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
					keyStr := string(key)
					valueStr := string(value)
					if messagesMap[keyStr] == nil {
						messagesMap[keyStr] = make(map[string]string)
					}
					messagesMap[keyStr][langCode] = valueStr
					return nil
				})
			}
		}
	}
	// 增加预置键
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
	// 关键字过滤
	lowerPrefix := strings.ToLower(query.Prefix)
	lowerDescription := strings.ToLower(query.Description)
	lowerLanguageCodes := strings.ToLower(query.LanguageCodes)
	lowerKeys := strings.ToLower(query.Keys)
	for k, values := range messagesMap {
		if len(values) == 0 {
			delete(messagesMap, k)
			continue
		}
		lowerKey := strings.ToLower(k)
		if query.Prefix != "" {
			if !strings.HasPrefix(lowerKey, lowerPrefix) {
				delete(messagesMap, k)
				continue
			}
		}
		if query.Keys != "" && !strings.Contains(lowerKeys, lowerKey) {
			delete(messagesMap, k)
			continue
		}
		if query.Description != "" {
			if !strings.HasPrefix(strings.ToLower(values["default"]), lowerDescription) {
				delete(messagesMap, k)
				continue
			}
		}
		for l := range values {
			if l == "default" {
				continue
			}
			if query.LanguageCodes != "" && !strings.Contains(lowerLanguageCodes, strings.ToLower(l)) {
				delete(values, l)
				continue
			}
		}
		messagesMap[k] = values
		if len(messagesMap[k]) == 0 {
			delete(messagesMap, k)
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
	_ = PublishCache(intlMessagesChangedChannel, langCode)
	return nil
}
