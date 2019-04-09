package pgo

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "reflect"
    "runtime"
    "strings"
    "sync"
)

// Application the pgo app,
// initialization steps:
// 1. import pgo: pgo.init()
// 2. customize app: optional
// 3. pgo.Run(): serve
//
// configuration:
// name:        "app-name"
// GOMAXPROCS:  2
// runtimePath: "@app/runtime"
// publicPath:  "@app/public"
// viewPath:    "@viewPath"
// container:   {}
// server:      {}
// components:  {}
type Application struct {
    mode        int    // running mode, WEB or CMD
    env         string // running env, eg. develop/online/testing-dev/testing-qa
    name        string // application name
    basePath    string // base path of application
    runtimePath string // runtime path for log etc.
    publicPath  string // public path for web assets
    viewPath    string // path for view template

    config     *Config
    container  *Container
    server     *Server
    components map[string]interface{}
    lock       sync.RWMutex
    router     *Router
    log        *Log
    status     *Status
    i18n       *I18n
    view       *View
    stopBefore *StopBefore // 服务停止前执行 [{"obj":"func"}]
}

func (app *Application) Construct() {
    exeBase := filepath.Base(os.Args[0])
    exeExt := filepath.Ext(os.Args[0])
    exeDir := filepath.Dir(os.Args[0])

    app.env = DefaultEnv
    app.mode = ModeWeb
    app.name = strings.TrimSuffix(exeBase, exeExt)
    app.basePath, _ = filepath.Abs(filepath.Join(exeDir, ".."))
    app.config = &Config{}
    app.container = &Container{}
    app.server = &Server{}
    app.components = make(map[string]interface{})
    app.stopBefore = &StopBefore{}
}

func (app *Application) Init() {
    env := flag.String("env", "", "set running env, eg. --env online")
    cmd := flag.String("cmd", "", "set running cmd, eg. --cmd /foo/bar")
    base := flag.String("base", "", "set base path, eg. --base /base/path")
    flag.Parse()

    // overwrite running env
    if len(*env) > 0 {
        app.env = *env
    }

    // overwrite running mode
    if len(*cmd) > 0 {
        app.mode = ModeCmd
    }

    // overwrite base path
    if len(*base) > 0 {
        app.basePath, _ = filepath.Abs(*base)
    }

    // set basic path alias
    type dummy struct{}
    pkgPath := reflect.TypeOf(dummy{}).PkgPath()
    SetAlias("@app", app.basePath)
    SetAlias("@pgo", strings.TrimPrefix(pkgPath, VendorPrefix))

    // initialize config object
    ConstructAndInit(app.config, nil)

    // initialize container object
    cntConf, _ := app.config.Get("app.container").(map[string]interface{})
    ConstructAndInit(app.container, cntConf)

    // initialize server object
    svrConf, _ := app.config.Get("app.server").(map[string]interface{})
    ConstructAndInit(app.server, svrConf)

    // overwrite app name
    if name := app.config.GetString("app.name", ""); len(name) > 0 {
        app.name = name
    }

    // overwrite GOMAXPROCS
    if n := app.config.GetInt("app.GOMAXPROCS", 0); n > 0 {
        runtime.GOMAXPROCS(n)
    }

    // set runtime path
    runtimePath := app.config.GetString("app.runtimePath", "@app/runtime")
    app.runtimePath, _ = filepath.Abs(GetAlias(runtimePath))
    SetAlias("@runtime", app.runtimePath)

    // set public path
    publicPath := app.config.GetString("app.publicPath", "@app/public")
    app.publicPath, _ = filepath.Abs(GetAlias(publicPath))
    SetAlias("@public", app.publicPath)

    // set view path
    viewPath := app.config.GetString("app.viewPath", "@app/view")
    app.viewPath, _ = filepath.Abs(GetAlias(viewPath))
    SetAlias("@view", app.viewPath)

    // set core components
    for id, class := range app.coreComponents() {
        key := fmt.Sprintf("app.components.%s.class", id)
        app.config.Set(key, class)
    }

    // create runtime directory if not exists
    if _, e := os.Stat(app.runtimePath); os.IsNotExist(e) {
        if e := os.MkdirAll(app.runtimePath, 0755); e != nil {
            panic(fmt.Sprintf("failed to create %s, %s", app.runtimePath, e))
        }
    }
}

// GetMode get running mode, web:1, cmd:2
func (app *Application) GetMode() int {
    return app.mode
}

// GetEnv get running env
func (app *Application) GetEnv() string {
    return app.env
}

// GetName get app name, default is executable name
func (app *Application) GetName() string {
    return app.name
}

// GetBasePath get base path, default is parent of executable
func (app *Application) GetBasePath() string {
    return app.basePath
}

// GetRuntimePath get runtime path, default is @app/runtime
func (app *Application) GetRuntimePath() string {
    return app.runtimePath
}

// GetPublicPath get public path, default is @app/public
func (app *Application) GetPublicPath() string {
    return app.publicPath
}

// GetViewPath get view path, default is @app/view
func (app *Application) GetViewPath() string {
    return app.viewPath
}

// GetConfig get config component
func (app *Application) GetConfig() *Config {
    return app.config
}

// GetContainer get container component
func (app *Application) GetContainer() *Container {
    return app.container
}

// GetServer get server component
func (app *Application) GetServer() *Server {
    return app.server
}

// GetRouter get router component
func (app *Application) GetRouter() *Router {
    if app.router == nil {
        app.router = app.Get("router").(*Router)
    }

    return app.router
}

// GetLog get log component
func (app *Application) GetLog() *Log {
    if app.log == nil {
        app.log = app.Get("log").(*Log)
    }

    return app.log
}

// GetStatus get status component
func (app *Application) GetStatus() *Status {
    if app.status == nil {
        app.status = app.Get("status").(*Status)
    }

    return app.status
}

// GetI18n get i18n component
func (app *Application) GetI18n() *I18n {
    if app.i18n == nil {
        app.i18n = app.Get("i18n").(*I18n)
    }

    return app.i18n
}

// GetView get view component
func (app *Application) GetView() *View {
    if app.view == nil {
        app.view = app.Get("view").(*View)
    }

    return app.view
}

// GetStopBefore get stopBefore component
func (app *Application) GetStopBefore() *StopBefore{
    return app.stopBefore
}

// Get get component by id
func (app *Application) Get(id string) interface{} {
    if _, ok := app.components[id]; !ok {
        app.loadComponent(id)
    }

    app.lock.RLock()
    defer app.lock.RUnlock()

    return app.components[id]
}

func (app *Application) loadComponent(id string) {
    app.lock.Lock()
    defer app.lock.Unlock()

    // avoid repeated loading
    if _, ok := app.components[id]; ok {
        return
    }

    conf := app.config.Get("app.components." + id)
    if conf == nil {
        panic("component not found: " + id)
    }

    app.components[id] = CreateObject(conf)
}

func (app *Application) coreComponents() map[string]string {
    return map[string]string{
        "router": "@pgo/Router",
        "log":    "@pgo/Log",
        "status": "@pgo/Status",
        "i18n":   "@pgo/I18n",
        "view":   "@pgo/View",
        "gzip":   "@pgo/Gzip",
        "file":   "@pgo/File",

        "http": "@pgo/Client/Http/Client",
    }
}
