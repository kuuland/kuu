package intl

import "sort"

// 语言代码规范：ISO 639-1（两位小写字母代码）
// 国家地区代码规范：ISO-3166-1（两位大写字母代码）
//
// 参考资料：
// https://zh.wikipedia.org/wiki/ISO_639-1%E4%BB%A3%E7%A0%81%E8%A1%A8
// https://zh.wikipedia.org/wiki/ISO_3166-1
// http://www.ruanyifeng.com/blog/2008/02/codes_for_language_names.html

var languageMap = map[string]string{
	// Arabic
	"ar": "العربية",
	// Azerbaijani
	"az": "Azərbaycanca",
	// English
	"en": "English",
	// Spanish
	"es": "Español",
	// French
	"fr": "Français",
	// Hungarian
	"hu": "magyar",
	// Italian
	"it": "Italiano",
	// Japanese
	"ja": "日本語",
	// Korean
	"ko": "한국어",
	// Mongolian
	"mn": "Монгол хэл",
	// Polish
	"pl": "Polski",
	// Portuguese (Brazil)
	"pt-BR": "Português do Brasil",
	// Russian
	"ru": "Русский",
	// Turkish
	"tr": "Türkçe",
	// Ukrainian
	"uk": "Українська",
	// Simplified Chinese
	"zh-Hans": "简体中文",
	// Traditional Chinese
	"zh-Hant": "繁體中文",
}

type Language struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func LanguageMap() map[string]string {
	return languageMap
}

func LanguageList() []Language {
	var keys []string
	for k := range languageMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var list []Language
	for _, k := range keys {
		list = append(list, Language{
			Code: k,
			Name: languageMap[k],
		})
	}
	return list
}

func ConvertLanguageCode(lang string) string {
	switch lang {
	case "zh", "zh-cn", "zh-CN":
		lang = "zh-Hans"
	case "zh-TW":
		lang = "zh-Hant"
	case "en-US":
		lang = "en"
	}
	return lang
}
