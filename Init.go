package pgo

import (
    "fmt"
    "reflect"
    "regexp"
    "strings"
    "time"

    "github.com/pinguo/pgo/Util"
)

const (
    ModeWeb            = 1
    ModeCmd            = 2
    DefaultEnv         = "develop"
    DefaultController  = "Index"
    DefaultAction      = "Index"
    DefaultHttpAddr    = "0.0.0.0:8000"
    DefaultTimeout     = 30 * time.Second
    DefaultHeaderBytes = 1 << 20
    ControllerWeb      = "Controller"
    ControllerCmd      = "Command"
    ConstructMethod    = "Construct"
    InitMethod         = "Init"
    VendorPrefix       = "vendor/"
    VendorLength       = 7
    ActionPrefix       = "Action"
    ActionLength       = 6
    TraceMaxDepth      = 10
    MaxPlugins         = 32
    MaxCacheObjects    = 100
)

var (
    App         = &Application{}
    appTime     = time.Now()
    aliasMap    = make(map[string]string)
    aliasRe     = regexp.MustCompile(`^@[^\\/]+`)
    logger      *Logger
    EmptyObject struct{}
)

// Map alias for map[string]interface{}
type Map map[string]interface{}

func init() {
    // initialize app
    ConstructAndInit(App, nil)

    // bind core object
    App.container.Bind(&Router{})
    App.container.Bind(&Log{})
    App.container.Bind(&ConsoleTarget{})
    App.container.Bind(&FileTarget{})
    App.container.Bind(&Status{})
    App.container.Bind(&I18n{})
    App.container.Bind(&View{})
    App.container.Bind(&Gzip{})
    App.container.Bind(&File{})
}

// Run run app
func Run() {
    App.GetServer().Serve()
}

// GLogger get global logger
func GLogger() *Logger {
    if logger == nil {
        // defer creation to first call, give opportunity to customize log target
        logger = App.GetLog().GetLogger(App.name, Util.GenUniqueId())
    }

    return logger
}

// TimeRun time duration since app run
func TimeRun() time.Duration {
    d := time.Since(appTime)
    d -= d % time.Second
    return d
}

// SetAlias set path alias, eg. @app => /path/to/base
func SetAlias(alias, path string) {
    if len(alias) > 0 && alias[0] != '@' {
        alias = "@" + alias
    }

    if strings.IndexAny(alias, `\/`) != -1 {
        panic("SetAlias: invalid alias, " + alias)
    }

    if len(alias) <= 1 || len(path) == 0 {
        panic("SetAlias: empty alias or path, " + alias)
    }

    aliasMap[alias] = path
}

// GetAlias resolve path alias, eg. @runtime/app.log => /path/to/runtime/app.log
func GetAlias(alias string) string {
    if prefix := aliasRe.FindString(alias); len(prefix) == 0 {
        return alias // not an alias
    } else if path, ok := aliasMap[prefix]; ok {
        return strings.Replace(alias, prefix, path, 1)
    }

    return ""
}

// CreateObject create object using the given configuration,
// class can be a string or a map contain "class" field,
// if a map is specified, fields except "class" will be
// treated as properties of the object to be created,
// params is optional parameters for Construct method.
func CreateObject(class interface{}, params ...interface{}) interface{} {
    var className string
    var config map[string]interface{}

    switch v := class.(type) {
    case string:
        className = v
    case map[string]interface{}:
        if _, ok := v["class"]; !ok {
            panic(`CreateObject: class configuration require "class" field`)
        }

        className = v["class"].(string)
        config = v
    default:
        panic(fmt.Sprintf("CreateObject: unsupported class type: %T", class))
    }

    if name := GetAlias(className); len(name) > 0 {
        return App.GetContainer().Get(name, config, params...).Interface()
    }

    panic("unknown class: " + className)
}

// Configure configure object using the given configuration,
// obj is a pointer or reflect.Value of a pointer,
// config is the configuration map for properties.
func Configure(obj interface{}, config map[string]interface{}) {
    // skip empty configuration
    if n := len(config); n == 0 {
        return
    } else if n == 1 {
        if _, ok := config["class"]; ok {
            return
        }
    }

    // v refer to the object pointer
    var v reflect.Value
    if _, ok := obj.(reflect.Value); ok {
        v = obj.(reflect.Value)
    } else {
        v = reflect.ValueOf(obj)
    }

    if v.Kind() != reflect.Ptr {
        panic("Configure: obj require a pointer or reflect.Value of a pointer")
    }

    // rv refer to the value of pointer
    rv := v.Elem()

    for key, val := range config {
        if key == "class" {
            continue
        }

        // change key to title string
        key = strings.Title(key)

        // check object's setter method
        if method := v.MethodByName("Set" + key); method.IsValid() {
            newVal := reflect.ValueOf(val).Convert(method.Type().In(0))
            method.Call([]reflect.Value{newVal})
            continue
        }

        // check object's public field
        field := rv.FieldByName(key)
        if field.IsValid() && field.CanSet() {
            newVal := reflect.ValueOf(val).Convert(field.Type())
            field.Set(newVal)
            continue
        }
    }
}

// ConstructAndInit construct and initialize object,
// obj is a pointer or reflect.Value of a pointer,
// config is configuration map for properties,
// params is optional parameters for Construct method.
func ConstructAndInit(obj interface{}, config map[string]interface{}, params ...interface{}) {
    var v reflect.Value
    if _, ok := obj.(reflect.Value); ok {
        v = obj.(reflect.Value)
    } else {
        v = reflect.ValueOf(obj)
    }

    if v.Kind() != reflect.Ptr {
        panic("ConstructAndInit: obj require a pointer or reflect.Value of a pointer")
    }

    // call Construct method
    if cm := v.MethodByName(ConstructMethod); cm.IsValid() {
        in := make([]reflect.Value, 0)
        for _, arg := range params {
            in = append(in, reflect.ValueOf(arg))
        }

        cm.Call(in)
    }

    // configure the object
    Configure(v, config)

    // call Init method
    if im := v.MethodByName(InitMethod); im.IsValid() {
        in := make([]reflect.Value, 0)
        im.Call(in)
    }
}
