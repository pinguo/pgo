package pgo

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    _ "net/http/pprof"
    "os"
    "os/signal"
    "path/filepath"
    "reflect"
    "runtime"
    "strings"
    "sync"
    "sync/atomic"
    "syscall"
    "time"
)

// Server the server component, configuration:
// server:
//     httpAddr:  "0.0.0.0:8000"
//     debugAddr: "0.0.0.0:8100"
//     httpsAddr: "0.0.0.0:8443"
//     crtFile: "@app/conf/site.crt"
//     keyFile: "@app/conf/site.key"
//     maxHeaderBytes: 1048576
//     readTimeout:   "30s"
//     writeTimeout:  "30s"
//     statsInterval: "60s"
//     enableAccessLog: true
//     maxPostBodySize: 1048576
type Server struct {
    httpAddr  string // address for http
    httpsAddr string // address for https
    debugAddr string // address for pprof

    crtFile         string        // https certificate file
    keyFile         string        // https private key file
    maxHeaderBytes  int           // max http header bytes
    readTimeout     time.Duration // timeout for reading request
    writeTimeout    time.Duration // timeout for writing response
    statsInterval   time.Duration // interval for output server stats
    enableAccessLog bool
    pluginNames     []string

    numReq  uint64         // request num handled
    plugins []IPlugin      // server plugin list
    servers []*http.Server // http server list
    pool    sync.Pool      // context pool
    maxPostBodySize int64  // max post body size
}

func (s *Server) Construct() {
    s.maxHeaderBytes = DefaultHeaderBytes
    s.readTimeout = DefaultTimeout
    s.writeTimeout = DefaultTimeout
    s.statsInterval = 60 * time.Second
    s.enableAccessLog = true
    s.pluginNames = []string{"gzip"}
    s.pool.New = func() interface{} {
        return new(Context)
    }
}

// SetHttpAddr set http addr, if both httpAddr and httpsAddr
// are empty, "0.0.0.0:8000" will be used as httpAddr.
func (s *Server) SetHttpAddr(addr string) {
    s.httpAddr = addr
}

// SetHttpsAddr set https addr.
func (s *Server) SetHttpsAddr(addr string) {
    s.httpsAddr = addr
}

// SetDebugAddr set debug and pprof addr.
func (s *Server) SetDebugAddr(addr string) {
    s.debugAddr = addr
}

// SetCrtFile set certificate file for https
func (s *Server) SetCrtFile(certFile string) {
    s.crtFile, _ = filepath.Abs(GetAlias(certFile))
}

// SetKeyFile set private key file for https
func (s *Server) SetKeyFile(keyFile string) {
    s.keyFile, _ = filepath.Abs(GetAlias(keyFile))
}

// SetMaxHeaderBytes set max header bytes
func (s *Server) SetMaxHeaderBytes(maxBytes int) {
    s.maxHeaderBytes = maxBytes
}

// SetMaxPostBodySize set max header bytes
func (s *Server) SetMaxPostBodySize(maxBytes int64) {
   s.maxPostBodySize = maxBytes
}

// SetReadTimeout set timeout to read request
func (s *Server) SetReadTimeout(v string) {
    if timeout, err := time.ParseDuration(v); err != nil {
        panic(fmt.Sprintf("Server: SetReadTimeout failed, val:%s, err:%s", v, err.Error()))
    } else {
        s.readTimeout = timeout
    }
}

// SetWriteTimeout set timeout to write response
func (s *Server) SetWriteTimeout(v string) {
    if timeout, err := time.ParseDuration(v); err != nil {
        panic(fmt.Sprintf("Server: SetWriteTimeout failed, val:%s, err:%s", v, err.Error()))
    } else {
        s.writeTimeout = timeout
    }
}

// SetStatsInterval set interval to output stats
func (s *Server) SetStatsInterval(v string) {
    if interval, err := time.ParseDuration(v); err != nil {
        panic(fmt.Sprintf("Server: SetStatsInterval failed, val:%s, err:%s", v, err.Error()))
    } else {
        s.statsInterval = interval
    }
}

// SetEnableAccessLog set access log enable or not
func (s *Server) SetEnableAccessLog(v bool) {
    s.enableAccessLog = v
}

// SetPlugins set plugin names
func (s *Server) SetPlugins(v []interface{}) {
    s.pluginNames = nil
    for _, vv := range v {
        s.pluginNames = append(s.pluginNames, vv.(string))
    }
}

// ServerStats server stats
type ServerStats struct {
    MemMB   uint   // memory obtained from os
    NumReq  uint64 // number of handled requests
    NumGO   uint   // number of goroutines
    NumGC   uint   // number of gc runs
    TimeGC  string // total time of gc pause
    TimeRun string // total time of app runs
}

// GetStats get server stats
func (s *Server) GetStats() *ServerStats {
    memStats := runtime.MemStats{}
    runtime.ReadMemStats(&memStats)

    timeGC := time.Duration(memStats.PauseTotalNs)
    if timeGC > time.Minute {
        timeGC -= timeGC % time.Second
    } else {
        timeGC -= timeGC % time.Millisecond
    }

    return &ServerStats{
        MemMB:   uint(memStats.Sys / (1 << 20)),
        NumReq:  atomic.LoadUint64(&s.numReq),
        NumGO:   uint(runtime.NumGoroutine()),
        NumGC:   uint(memStats.NumGC),
        TimeGC:  timeGC.String(),
        TimeRun: TimeRun().String(),
    }
}

// Serve request processing entry
func (s *Server) Serve() {
    // flush log when app end
    defer App.GetLog().Flush()
    // exec stopBefore when app end
    defer App.GetStopBefore().Exec()

    // initialize plugins
    s.initPlugins()

    // process command request
    if App.GetMode() == ModeCmd {
        s.ServeCMD()
        return
    }

    // process http request
    if s.httpAddr == "" && s.httpsAddr == "" {
        s.httpAddr = DefaultHttpAddr
    }

    wg := sync.WaitGroup{}
    s.handleHttp(&wg)
    s.handleHttps(&wg)
    s.handleDebug(&wg)
    s.handleSignal(&wg)
    s.handleStats(&wg)
    wg.Wait()
}

// ServeCMD serve command request
func (s *Server) ServeCMD() {
    ctx := Context{server: s}

    // only apply the last plugin for command
    ctx.process(s.plugins[len(s.plugins)-1:])
}

// ServeHTTP serve http request
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Change the maxPostBodySize
    if s.maxPostBodySize > 0 {
        r.Body = http.MaxBytesReader(w, r.Body, s.maxPostBodySize)
    }
    // increase request num
    atomic.AddUint64(&s.numReq, 1)
    ctx := s.pool.Get().(*Context)

    ctx.server = s
    ctx.input = r
    ctx.output = &ctx.response
    ctx.response.reset(w)
    ctx.process(s.plugins)
    s.pool.Put(ctx)
}

// HandleRequest handle request of cmd or http,
// this method called in the last of plugin chain.
func (s *Server) HandleRequest(ctx *Context) {
    // get request path and resolve route
    path := ctx.GetPath()
    route, params := App.GetRouter().Resolve(path)

    // get new controller bind to this route
    rv, action := s.createController(route, ctx)
    if !rv.IsValid() {
        ctx.End(http.StatusNotFound, []byte("route not found"))
        return
    }

    actionId := ctx.GetActionId()
    controller := rv.Interface().(IController)

    // fill empty string for missing param
    numIn := action.Type().NumIn()
    if len(params) < numIn {
        fill := make([]string, numIn-len(params))
        params = append(params, fill...)
    }

    // prepare params for action call
    callParams := make([]reflect.Value, 0)
    for _, param := range params {
        callParams = append(callParams, reflect.ValueOf(param))
    }

    defer func() {
        // process controller panic
        if v := recover(); v != nil {
            controller.HandlePanic(v)
        }

        // after action hook
        controller.AfterAction(actionId)
    }()

    // before action hook
    controller.BeforeAction(actionId)

    // call action method
    action.Call(callParams)
}

func (s *Server) handleHttp(wg *sync.WaitGroup) {
    if s.httpAddr == "" {
        return
    }

    svr := s.newHttpServer(s.httpAddr)
    s.servers = append(s.servers, svr)
    wg.Add(1)

    GLogger().Info("start running http at " + svr.Addr)

    go func() {
        if err := svr.ListenAndServe(); err != http.ErrServerClosed {
            panic("ListenAndServe failed, " + err.Error())
        }
    }()
}

func (s *Server) handleHttps(wg *sync.WaitGroup) {
    if s.httpsAddr == "" {
        return
    } else if s.crtFile == "" || s.keyFile == "" {
        panic("https no crtFile or keyFile configured")
    }

    svr := s.newHttpServer(s.httpsAddr)
    s.servers = append(s.servers, svr)
    wg.Add(1)

    GLogger().Info("start running https at " + svr.Addr)

    go func() {
        if err := svr.ListenAndServeTLS(s.crtFile, s.keyFile); err != http.ErrServerClosed {
            panic("ListenAndServeTLS failed, " + err.Error())
        }
    }()
}

func (s *Server) handleDebug(wg *sync.WaitGroup) {
    if s.debugAddr == "" {
        return
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        data, _ := json.Marshal(s.GetStats())
        w.Write(data)
    })

    svr := s.newHttpServer(s.debugAddr)
    svr.Handler = nil // use default handler
    s.servers = append(s.servers, svr)
    wg.Add(1)

    GLogger().Info("start running debug at " + svr.Addr)

    go func() {
        if err := svr.ListenAndServe(); err != http.ErrServerClosed {
            panic("ListenAndServe failed, " + err.Error())
        }
    }()
}

func (s *Server) handleSignal(wg *sync.WaitGroup) {
    sig := make(chan os.Signal)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sig // wait signal
        for _, svr := range s.servers {
            GLogger().Info("stop running " + svr.Addr)
            ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
            svr.Shutdown(ctx)
            wg.Done()
        }
    }()
}

func (s *Server) handleStats(wg *sync.WaitGroup) {
    timer := time.Tick(s.statsInterval)

    go func() {
        for {
            <-timer // wait timer
            data, _ := json.Marshal(s.GetStats())
            GLogger().Info("app stats: " + string(data))
        }
    }()
}

func (s *Server) newHttpServer(addr string) *http.Server {
    return &http.Server{
        Addr:           addr,
        ReadTimeout:    s.readTimeout,
        WriteTimeout:   s.writeTimeout,
        MaxHeaderBytes: s.maxHeaderBytes,
        Handler:        s,
    }
}

func (s *Server) initPlugins() {
    for _, name := range s.pluginNames {
        s.plugins = append(s.plugins, App.Get(name).(IPlugin))
    }

    // server is the last plugin
    s.plugins = append(s.plugins, s)

    if len(s.plugins) > MaxPlugins {
        panic("Server: too many plugins")
    }
}

func (s *Server) createController(route string, ctx *Context) (reflect.Value, reflect.Value) {
    if "/" == route {
        route += DefaultController
    }

    var controllerId, actionId string
    container := App.GetContainer()

    pos := strings.LastIndexByte(route, '/')
    if pos > 0 && container.Has(s.getControllerName(route[:pos])) {
        controllerId = route[:pos]
        actionId = route[pos+1:]
    } else if container.Has(s.getControllerName(route)) {
        controllerId = route
        actionId = ""
    } else {
        return reflect.Value{}, reflect.Value{}
    }

    controllerName := s.getControllerName(controllerId)
    actions, _ := container.GetInfo(controllerName).(map[string]int)

    if len(actionId) > 0 {
        if _, ok := actions[actionId]; !ok {
            return reflect.Value{}, reflect.Value{}
        }
    } else {
        if _, ok := actions[DefaultAction]; ok {
            actionId = DefaultAction
        } else {
            method := ctx.GetMethod()
            if _, ok := actions[method]; !ok {
                return reflect.Value{}, reflect.Value{}
            }
            actionId = method
        }
    }

    ctx.SetControllerId(controllerId)
    ctx.SetActionId(actionId)

    controller := container.Get(controllerName, nil, ctx)
    action := controller.Method(actions[actionId])
    return controller, action
}

func (s *Server) getControllerName(id string) string {
    if ModeWeb == App.mode {
        return ControllerWeb + id + ControllerWeb
    } else {
        return ControllerCmd + id + ControllerCmd
    }
}
