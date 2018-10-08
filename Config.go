package pgo

import (
    "encoding/json"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "sync"

    "github.com/pinguo/pgo/Util"
)

type Config struct {
    parsers map[string]IConfigParser
    data    map[string]interface{}
    paths   []string
    lock    sync.RWMutex
}

func (c *Config) Construct() {
    c.parsers = make(map[string]IConfigParser)
    c.data = make(map[string]interface{})
    c.paths = make([]string, 0)

    confPath := filepath.Join(App.GetBasePath(), "conf")
    if f, e := os.Stat(confPath); os.IsNotExist(e) || !f.IsDir() {
        panic("invalid config path, " + confPath)
    }

    c.AddPath(confPath)
    c.AddPath(filepath.Join(confPath, App.GetEnv()))

    c.AddParser("json", &JsonConfigParser{})
}

// add parser for file with ext extension
func (c *Config) AddParser(ext string, parser IConfigParser) {
    c.parsers[ext] = parser
}

// add path to end of search paths
func (c *Config) AddPath(path string) {
    paths := make([]string, 0)
    for _, v := range c.paths {
        if v != path {
            paths = append(paths, v)
        }
    }

    c.paths = append(paths, path)
}

func (c *Config) GetBool(key string, dft bool) bool {
    if v := c.Get(key); v != nil {
        return Util.ToBool(v)
    }

    return dft
}

func (c *Config) GetInt(key string, dft int) int {
    if v := c.Get(key); v != nil {
        return Util.ToInt(v)
    }

    return dft
}

func (c *Config) GetFloat(key string, dft float64) float64 {
    if v := c.Get(key); v != nil {
        return Util.ToFloat(v)
    }

    return dft
}

func (c *Config) GetString(key string, dft string) string {
    if v := c.Get(key); v != nil {
        return Util.ToString(v)
    }

    return dft
}

// get config by dot separated key, empty key for all loaded config
func (c *Config) Get(key string) interface{} {
    ks := strings.Split(key, ".")
    if _, ok := c.data[ks[0]]; !ok {
        c.Load(ks[0])
    }

    c.lock.RLock()
    defer c.lock.RUnlock()

    return Util.MapGet(c.data, key)
}

// set config by dot separated key, empty key for root, nil val for clear
func (c *Config) Set(key string, val interface{}) {
    c.lock.Lock()
    defer c.lock.Unlock()

    Util.MapSet(c.data, key, val)
}

// load config file
func (c *Config) Load(name string) {
    c.lock.Lock()
    defer c.lock.Unlock()

    // avoid repeated loading
    _, ok := c.data[name]
    if ok || len(name) == 0 {
        return
    }

    for _, path := range c.paths {
        files, _ := filepath.Glob(filepath.Join(path, name+".*"))
        for _, f := range files {
            ext := strings.ToLower(filepath.Ext(f))
            if parser, ok := c.parsers[ext[1:]]; ok {
                if conf := parser.Parse(f); conf != nil {
                    Util.MapMerge(c.data, map[string]interface{}{name: conf})
                } else {
                    panic("Config: failed to parse file: " + f)
                }
            }
        }
    }
}

// parser for json config
type JsonConfigParser struct {
}

func (j *JsonConfigParser) Parse(path string) map[string]interface{} {
    h, e := os.Open(path)
    if e != nil {
        panic("JsonConfigParser: failed to open file: " + path)
    }

    defer h.Close()

    content, e := ioutil.ReadAll(h)
    if e != nil {
        panic("JsonConfigParser: failed to read file: " + path)
    }

    // expand env: ${env||default}
    content = Util.ExpandEnv(content)

    var data map[string]interface{}
    if e := json.Unmarshal(content, &data); e != nil {
        panic("jsonConfigParser: failed to parse file: " + path)
    }

    return data
}
