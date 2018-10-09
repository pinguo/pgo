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
    DefaultEnv         = "prod"
    DefaultController  = "Index"
    DefaultAction      = "Index"
    DefaultServerAddr  = "0.0.0.0:8000"
    DefaultTimeout     = 30 * time.Second
    DefaultHeaderBytes = 1 << 20
    ControllerWeb      = "Controller"
    ControllerCmd      = "Command"
    ConstructMethod    = "Construct"
    InitMethod         = "Init"
    VendorPrefix       = "vendor/"
    VendorLength       = 7
    TraceMaxDepth      = 10
)

var (
    aliases map[string]string
    aliasRe *regexp.Regexp
    logger  *Logger

    App         *Application
    EmptyObject struct{}
)

// alias for map[string]interface
type Map map[string]interface{}

func init() {
    // global initialization
    aliases = make(map[string]string)
    aliasRe = regexp.MustCompile(`^@[^\\/]+`)

    // new app instance
    App = &Application{}
    ConstructAndInit(App, nil)

    // bind core object
    App.container.Bind(&Router{})
    App.container.Bind(&Dispatcher{})
    App.container.Bind(&ConsoleTarget{})
    App.container.Bind(&FileTarget{})
    App.container.Bind(&Status{})
    App.container.Bind(&I18n{})
    App.container.Bind(&View{})
}

// run application
func Run() {
    App.GetServer().Serve()
}

// get global logger
func GLogger() *Logger {
    if logger == nil {
        // defer creation to first call, give opportunity to customize log target
        logger = App.GetLog().GetLogger(App.name, Util.GenUniqueId())
    }

    return logger
}

// set alias for path, @app => /path/to/base
func SetAlias(alias, path string) {
    if len(alias) > 0 && alias[0] != '@' {
        alias = "@" + alias
    }

    if strings.IndexAny(alias, `\/`) != -1 {
        panic("SetAlias: invalid alias format, " + alias)
    }

    if len(alias) <= 1 || len(path) == 0 {
        panic("SetAlias: alias or path cannot be empty")
    }

    aliases[alias] = path
}

// resolve path alias, @runtime/app.log => /path/to/runtime/App.log
func GetAlias(alias string) string {
    rn := aliasRe.FindString(alias)

    // not an alias
    if len(rn) == 0 {
        return alias
    }

    if path, ok := aliases[rn]; ok {
        return strings.Replace(alias, rn, path, 1)
    }

    return ""
}

// create object using the given configuration
func CreateObject(class interface{}, params ...interface{}) interface{} {
    var className string
    var config map[string]interface{}

    switch v := class.(type) {
    case string:
        className = v
    case map[string]interface{}:
        if _, ok := v["class"]; !ok {
            panic(`CreateObject: object configuration require "class" element`)
        }

        className = v["class"].(string)
        config = v
    default:
        panic(fmt.Sprintf("CreateObject: unsupported class type: %T", class))
    }

    if name := GetAlias(className); len(name) > 0 {
        return App.GetContainer().Get(name, config, params...)
    }

    panic("unknown class: " + className)
}

// configure object using the given configuration
func Configure(obj interface{}, config map[string]interface{}) {
    // v refer to the object pointer
    var v reflect.Value
    if _, ok := obj.(reflect.Value); ok {
        v = obj.(reflect.Value)
    } else {
        v = reflect.ValueOf(obj)
    }

    if v.Kind() != reflect.Ptr {
        panic("Configure: obj require a pointer or a reflect.Value of pointer")
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
        if rv.Kind() == reflect.Struct {
            field := rv.FieldByName(key)
            if field.IsValid() && field.CanSet() {
                newVal := reflect.ValueOf(val).Convert(field.Type())
                field.Set(newVal)
                continue
            }
        }
    }
}

// construct and initialize object
func ConstructAndInit(obj interface{}, config map[string]interface{}, params ...interface{}) {
    var v reflect.Value
    if _, ok := obj.(reflect.Value); ok {
        v = obj.(reflect.Value)
    } else {
        v = reflect.ValueOf(obj)
    }

    if v.Kind() != reflect.Ptr {
        panic("ConstructAndInit: obj require a pointer or a reflect.Value of pointer")
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
