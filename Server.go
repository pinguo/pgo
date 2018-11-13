package pgo

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
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

    "github.com/pinguo/pgo/Util"
)

// ServerStats server stats
type ServerStats struct {
    TotalReq uint64
    MemMB    uint
    NumGO    uint
    NumGC    uint
    TimeGC   string
    TimeRun  string
}

// Server the server component, configuration:
// "server": {
//     "httpAddr": "0.0.0.0:8000",
//     "debugAddr": "0.0.0.0:8100",
//     "httpsAddr": "0.0.0.0:8443",
//     "certFile": "@app/conf/server.crt",
//     "keyFile": "@app/conf/server.key",
//     "readTimeout": "30s",
//     "writeTimeout": "30s",
//     "maxHeaderBytes": 1024000,
//     "statsInterval": "60s",
//     "errorLogOff": [404]
// }
type Server struct {
    httpAddr  string // address for http
    httpsAddr string // address for https
    debugAddr string // address for stats and pprof

    certFile       string // https certificate file
    keyFile        string // https private key file
    readTimeout    time.Duration
    writeTimeout   time.Duration
    maxHeaderBytes int

    plugins []IPlugin

    FileEnable   bool // static file in public path enabled
    GzipEnable   bool // gzip output compress enabled
    GzipMinBytes int  // minimum bytes for gzip output

    statsInterval time.Duration // interval for output server stats
    errorLogOff   map[int]bool  // close error log for specific code

    totalReq uint64 // total requests since server start
    servers  []*http.Server
}

func (s *Server) Construct() {
    s.readTimeout = DefaultTimeout
    s.writeTimeout = DefaultTimeout
    s.maxHeaderBytes = DefaultHeaderBytes

    s.FileEnable = true
    s.GzipEnable = true
    s.GzipMinBytes = 1024

    s.statsInterval = 60 * time.Second
}

// SetHttpAddr set http addr, default "0.0.0.0:8000"
func (s *Server) SetHttpAddr(addr string) {
    s.httpAddr = addr
}

// SetHttpsAddr set https addr, default ""
func (s *Server) SetHttpsAddr(addr string) {
    s.httpsAddr = addr
}

// SetDebugAddr set debug and pprof addr, default ""
func (s *Server) SetDebugAddr(addr string) {
    s.debugAddr = addr
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

// SetMaxHeaderBytes set max header bytes for http
func (s *Server) SetMaxHeaderBytes(maxBytes int) {
    s.maxHeaderBytes = maxBytes
}

// SetCertFile set certificate file for https
func (s *Server) SetCertFile(certFile string) {
    s.certFile, _ = filepath.Abs(GetAlias(certFile))
}

// SerKeyFile set private key file for https
func (s *Server) SerKeyFile(keyFile string) {
    s.keyFile, _ = filepath.Abs(GetAlias(keyFile))
}

func (s *Server) SetErrorLogOff(codes []interface{}) {
    s.errorLogOff = make(map[int]bool)
    for _, v := range codes {
        s.errorLogOff[Util.ToInt(v)] = true
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

func (s *Server) IsErrorLogOff(status int) bool {
    return s.errorLogOff[status]
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
        TotalReq: atomic.LoadUint64(&s.totalReq),
        MemMB:    uint(memStats.Sys / (1 << 20)),
        NumGO:    uint(runtime.NumGoroutine()),
        NumGC:    uint(memStats.NumGC),
        TimeGC:   timeGC.String(),
        TimeRun:  TimeRun().String(),
    }
}

// Serve entry of request processing
func (s *Server) Serve() {
    // flush log when app end
    defer App.GetLog().Flush()

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
    ctx := &Context{}
    ctx.Init()

    s.handleRequest(ctx)
}

// ServeHTTP serve http request
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    atomic.AddUint64(&s.totalReq, 1)

    if s.FileEnable {
        // process static file
        if ext := filepath.Ext(r.URL.Path); len(ext) > 0 {
            s.handleFile(w, r)
            return
        }
    }

    // process http service
    ctx := &Context{}
    ctx.SetInput(r)
    ctx.SetOutput(w)
    ctx.Init()
    s.handleRequest(ctx)
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
    } else if s.certFile == "" || s.keyFile == "" {
        panic("https no certFile or keyFile configured")
    }

    svr := s.newHttpServer(s.httpsAddr)
    s.servers = append(s.servers, svr)
    wg.Add(1)

    GLogger().Info("start running https at " + svr.Addr)

    go func() {
        if err := svr.ListenAndServeTLS(s.certFile, s.keyFile); err != http.ErrServerClosed {
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

// handle file in public path, no gzip support
func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet && r.Method != http.MethodHead {
        http.Error(w, "", http.StatusMethodNotAllowed)
        return
    }

    file := filepath.Join(App.GetPublicPath(), Util.CleanPath(r.URL.Path))
    h, e := os.Open(file)
    if e != nil {
        http.Error(w, "", http.StatusNotFound)
        return
    }

    defer h.Close()

    f, _ := h.Stat()
    http.ServeContent(w, r, file, f.ModTime(), h)
}

func (s *Server) handleRequest(ctx *Context) {
    defer func() {
        // process unhandled panic
        if v := recover(); v != nil {
            s.handlePanic(ctx, v)
        }
    }()

    // get request path and resolve route
    path := ctx.GetPath()
    route, params := App.GetRouter().Resolve(path)

    // get new controller bind to this route
    rv, info := s.createController(route, ctx)
    controller := rv.Interface().(IController)

    // get action method by sequence number
    actionMap := info.(map[string]int)
    actionId := ctx.GetActionId()
    action := rv.Method(actionMap[actionId])

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

        // send action output
        controller.FinishAction(actionId)
    }()

    // before action hook
    controller.BeforeAction(actionId)

    // call action method
    action.Call(callParams)

    // after action hook
    controller.AfterAction(actionId)
}

func (s *Server) createController(route string, ctx *Context) (reflect.Value, interface{}) {
    if "/" == route {
        route += DefaultController
    }

    var controllerId, actionId string
    di := App.GetContainer()

    pos := strings.LastIndexByte(route, '/')
    if pos > 0 && di.Has(s.getControllerName(route[:pos])) {
        controllerId = route[:pos]
        actionId = route[pos+1:]
    } else if di.Has(s.getControllerName(route)) {
        controllerId = route
        actionId = ""
    } else {
        panic(NewException(http.StatusNotFound, "route not found, %s", route))
    }

    rv, info := di.GetValue(s.getControllerName(controllerId), nil)
    actions := info.(map[string]int)

    if len(actionId) > 0 {
        if _, ok := actions[actionId]; !ok {
            panic(NewException(http.StatusNotFound, "route not found, %s", route))
        }
    } else {
        if _, ok := actions[DefaultAction]; ok {
            actionId = DefaultAction
        } else {
            method := ctx.GetMethod()
            if _, ok := actions[method]; !ok {
                panic(NewException(http.StatusNotFound, "route not found, %s", route))
            }
            actionId = method
        }
    }

    ctx.SetControllerId(controllerId)
    ctx.SetActionId(actionId)
    rv.Interface().(IObject).SetContext(ctx)

    return rv, info
}

func (s *Server) getControllerName(id string) string {
    if ModeWeb == App.mode {
        return ControllerWeb + id + ControllerWeb
    } else {
        return ControllerCmd + id + ControllerCmd
    }
}

func (s *Server) handlePanic(ctx *Context, v interface{}) {
    status := http.StatusInternalServerError
    switch e := v.(type) {
    case *Exception:
        status = e.GetStatus()
        ctx.End(status, []byte(App.GetStatus().GetText(status, ctx, e.GetMessage())))
    default:
        ctx.End(status, []byte(http.StatusText(status)))
    }

    if !s.IsErrorLogOff(status) {
        ctx.Error("%s, trace[%s]", Util.ToString(v), Util.PanicTrace(TraceMaxDepth, false))
    }
}
