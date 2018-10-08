package pgo

import (
    "fmt"
    "pgo/Util"
    "strings"
)

// i18n component, lang format: ll-CC or ll
// lower case lang code, upper case area code
// lang file located in conf directory with
// name format: i18n_{lang}.json
// configuration:
// {
//     "sourceLang": "en",
//     "targetLang": [ "en", "zh-CN", "zh-TW"]
// }
type I18n struct {
    sourceLang string
    targetLang map[string]bool
}

func (i *I18n) Construct() {
    i.sourceLang = "en"
    i.targetLang = make(map[string]bool)
}

func (i *I18n) SetSourceLang(lang string) {
    i.sourceLang = lang
}

func (i *I18n) SetTargetLang(targets []interface{}) {
    for _, v := range targets {
        lang := Util.ToString(v)
        i.targetLang[lang] = true
    }
}

// translate message to target lang, lang format is one of the following:
// 1. accept-language header value: zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7
// 2. ll-CC: lower case lang code and upper case area code, zh-CN
// 3. ll: lower case of lang code without area code, zh
func (i *I18n) Translate(message, lang string, params ...interface{}) string {
    translation := i.loadMessage(message, i.detectLang(lang))
    if len(params) > 0 {
        return fmt.Sprintf(translation, params...)
    }

    return translation
}

// detect support lang, lang can be accept-language header
func (i *I18n) detectLang(lang string) string {
    // use first part of accept-language
    if pos := strings.IndexByte(lang, ','); pos > 0 {
        lang = lang[:pos]
    }

    // omit q weight
    if pos := strings.IndexByte(lang, ';'); pos > 0 {
        lang = lang[:pos]
    }

    // format lang to ll-CC format
    lang = Util.FormatLanguage(lang)

    if i.targetLang[lang] {
        return lang
    }

    if pos := strings.IndexByte(lang, '-'); pos > 0 {
        if lang = lang[:pos]; i.targetLang[lang] {
            return lang
        }
    }

    return i.sourceLang
}

// load message from lang file i18n_{lang}.json
func (i *I18n) loadMessage(message, lang string) string {
    if !i.targetLang[lang] {
        return message
    }

    key := fmt.Sprintf("i18n_%s.%s", lang, message)
    return App.GetConfig().GetString(key, message)
}
