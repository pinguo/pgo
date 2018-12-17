package MaxMind

import (
    "fmt"
    "net"
    "strings"

    "github.com/oschwald/maxminddb-golang"
    "github.com/pinguo/pgo"
)

// Geo lookup result
type Geo struct {
    Code      string `json:"code"`               // country/area code
    Continent string `json:"-"`                  // continent name (en)
    Country   string `json:"country,omitempty"`  // country/area name (en)
    Province  string `json:"province,omitempty"` // province name (en)
    City      string `json:"city,omitempty"`     // city name(en)

    // i18n name, default is en
    I18n struct {
        Continent string
        Country   string
        Province  string
        City      string
    } `json:"-"`
}

// MaxMind Client component, configuration:
// maxMind:
//     class: "@pgo/Client/MaxMind/Client"
//     countryFile: "@app/../geoip/GeoLite2-Country.mmdb"
//     cityFile: "@app/../geoip/GeoLite2-City.mmdb"
type Client struct {
    readers [2]*maxminddb.Reader
}

func (c *Client) Init() {
    if c.readers[DBCountry] == nil && c.readers[DBCity] == nil {
        panic("MaxMind: both country and city db are empty")
    }
}

func (c *Client) SetCountryFile(path string) {
    c.loadFile(DBCountry, path)
}

func (c *Client) SetCityFile(path string) {
    c.loadFile(DBCity, path)
}

// get geo info by ip, optional args:
// db int: preferred geo db
// lang string: preferred i18n language
func (c *Client) GeoByIp(ip string, args ...interface{}) *Geo {
    db := DBCity
    lang := defaultLang

    // parse optional args
    for _, arg := range args {
        switch v := arg.(type) {
        case int:
            db = v
        case string:
            lang = v
        default:
            panic(fmt.Sprintf("MaxMind: invalid arg type: %T", arg))
        }
    }

    // get available db reader
    if c.readers[db] == nil {
        db = (db + 1) % 2
    }

    var m map[string]interface{}
    if e := c.readers[db].Lookup(net.ParseIP(ip), &m); e != nil {
        panic(fmt.Sprintf("MaxMind: failed to parse ip, ip:%s, err:%s", ip, e))
    }

    if len(m) == 0 {
        return nil
    }

    geo := &Geo{}
    for k, v := range m {
        switch k {
        case "continent":
            vm, _ := v.(map[string]interface{})
            geo.Continent = c.getI18nName(vm, defaultLang)
            geo.I18n.Continent = c.getI18nName(vm, lang)
        case "country":
            vm, _ := v.(map[string]interface{})
            geo.Code = vm["iso_code"].(string)
            geo.Country = c.getI18nName(vm, defaultLang)
            geo.I18n.Country = c.getI18nName(vm, lang)
        case "city":
            vm, _ := v.(map[string]interface{})
            geo.City = c.getI18nName(vm, defaultLang)
            geo.I18n.City = c.getI18nName(vm, lang)
        case "subdivisions":
            if vs, _ := v.([]interface{}); len(vs) > 0 {
                vm, _ := vs[0].(map[string]interface{})
                geo.Province = c.getI18nName(vm, defaultLang)
                geo.I18n.Province = c.getI18nName(vm, lang)
            }
        }
    }

    return geo
}

func (c *Client) loadFile(db int, path string) {
    if reader, e := maxminddb.Open(pgo.GetAlias(path)); e != nil {
        panic(fmt.Sprintf("MaxMind: failed to open file, path:%s, err:%s", path, e))
    } else {
        c.readers[db] = reader
    }
}

func (c *Client) getI18nName(m map[string]interface{}, lang string) string {
    names, _ := m["names"].(map[string]interface{})

    if n, ok := names[lang]; ok {
        return n.(string)
    } else if p := strings.IndexAny(lang, "_-"); p > 0 {
        l := lang[:p]
        if n, ok := names[l]; ok {
            return n.(string)
        }
    }

    if n, ok := names[defaultLang]; ok {
        return n.(string)
    }

    return ""
}
