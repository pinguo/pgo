package pgo

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "sync"

    "github.com/pinguo/pgo/Util"
)

// Config the config component
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
    if f, _ := os.Stat(confPath); f != nil && f.IsDir() {
        c.paths = append(c.paths, confPath)
    } else {
        panic("Config: invalid conf path, " + confPath)
    }

    envPath := filepath.Join(confPath, App.GetEnv())
    if f, _ := os.Stat(envPath); f != nil && f.IsDir() {
        c.paths = append(c.paths, envPath)
    } else if App.GetEnv() != DefaultEnv {
        panic("Config: invalid env path, " + envPath)
    }

    c.AddParser("json", &JsonConfigParser{})
    c.AddParser("yaml", &YamlConfigParser{})
}

// AddParser add parser for file with ext extension
func (c *Config) AddParser(ext string, parser IConfigParser) {
    c.parsers[ext] = parser
}

// AddPath add path to end of search paths
func (c *Config) AddPath(path string) {
    paths := make([]string, 0)
    for _, v := range c.paths {
        if v != path {
            paths = append(paths, v)
        }
    }

    c.paths = append(paths, path)
}

// GetBool get bool value from config,
// key is dot separated config key,
// dft is default value if key not exists.
func (c *Config) GetBool(key string, dft bool) bool {
    if v := c.Get(key); v != nil {
        return Util.ToBool(v)
    }

    return dft
}

// GetInt get int value from config,
// key is dot separated config key,
// dft is default value if key not exists.
func (c *Config) GetInt(key string, dft int) int {
    if v := c.Get(key); v != nil {
        return Util.ToInt(v)
    }

    return dft
}

// GetFloat get float value from config,
// key is dot separated config key,
// dft is default value if key not exists.
func (c *Config) GetFloat(key string, dft float64) float64 {
    if v := c.Get(key); v != nil {
        return Util.ToFloat(v)
    }

    return dft
}

// GetString get string value from config,
// key is dot separated config key,
// dft is default value if key not exists.
func (c *Config) GetString(key string, dft string) string {
    if v := c.Get(key); v != nil {
        return Util.ToString(v)
    }

    return dft
}

// GetSliceBool get []bool value from config,
// key is dot separated config key,
// nil is default value if key not exists.
func (c *Config) GetSliceBool(key string) []bool {
    var ret []bool
    if v := c.Get(key); v != nil {
        if vI, ok := v.([]interface{}); ok == true {
            for _, vv := range vI {
                ret = append(ret, Util.ToBool(vv))
            }
        }
    }

    return ret
}

// GetSliceInt get []int value from config,
// key is dot separated config key,
// nil is default value if key not exists.
func (c *Config) GetSliceInt(key string) []int {
    var ret []int
    if v := c.Get(key); v != nil {
        if vI, ok := v.([]interface{}); ok == true {
            for _, vv := range vI {
                ret = append(ret, Util.ToInt(vv))
            }
        }
    }

    return ret
}

// GetSliceFloat get []float value from config,
// key is dot separated config key,
// nil is default value if key not exists.
func (c *Config) GetSliceFloat(key string) []float64 {
    var ret []float64
    if v := c.Get(key); v != nil {
        if vI, ok := v.([]interface{}); ok == true {
            for _, vv := range vI {
                ret = append(ret, Util.ToFloat(vv))
            }
        }
    }

    return ret
}

// GetSliceString get []string value from config,
// key is dot separated config key,
// nil is default value if key not exists.
func (c *Config) GetSliceString(key string) []string {
    var ret []string
    if v := c.Get(key); v != nil {
        if vI, ok := v.([]interface{}); ok == true {
            for _, vv := range vI {
                ret = append(ret, Util.ToString(vv))
            }
        }
    }

    return ret
}

// Get get value by dot separated key,
// the first part of key is file name
// without extension. if key is empty,
// all loaded config will be returned.
func (c *Config) Get(key string) interface{} {
    ks := strings.Split(key, ".")
    if _, ok := c.data[ks[0]]; !ok {
        c.Load(ks[0])
    }

    c.lock.RLock()
    defer c.lock.RUnlock()

    return Util.MapGet(c.data, key)
}

// Set set value by dot separated key,
// if key is empty, the value will set
// to root, if val is nil, the key will
// be deleted.
func (c *Config) Set(key string, val interface{}) {
    c.lock.Lock()
    defer c.lock.Unlock()

    Util.MapSet(c.data, key, val)
}

// Load load config file under the search paths.
// file under env sub path will be merged.
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
                }
            }
        }
    }
}

// JsonConfigParser parser for json config
type JsonConfigParser struct {
}

// Parse parse json config, environment value like ${env||default} will expand
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
        panic(fmt.Sprintf("jsonConfigParser: failed to parse file: %s, %s", path, e.Error()))
    }

    return data
}

// YamlConfigParser parser for yaml config
type YamlConfigParser struct {
}

// Parse parse yaml config, environment value like ${env||default} will expand
func (y *YamlConfigParser) Parse(path string) map[string]interface{} {
    h, e := os.Open(path)
    if e != nil {
        panic("YamlConfigParser: failed to open file: " + path)
    }

    defer h.Close()

    content, e := ioutil.ReadAll(h)
    if e != nil {
        panic("YamlConfigParser: failed to read file: " + path)
    }

    // expand env: ${env||default}
    content = Util.ExpandEnv(content)

    var data map[string]interface{}
    if e := Util.YamlUnmarshal(content, &data); e != nil {
        panic(fmt.Sprintf("YamlConfigParser: failed to parse file: %s, %s", path, e.Error()))
    }

    return data
}
