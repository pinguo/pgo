package pgo

import (
    "context"
    "flag"
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

// server configuration:
// "server": {
//     "addr": "0.0.0.0:8000",
//     "readTimeout": "30s",
//     "writeTimeout": "30s",
//     "maxHeaderBytes": 1024000,
//     "fileEnable": true,
//     "gzipEnable": true,
//     "gzipMinBytes": 1024,
//     "statsInterval": "60s",
//     "errorLogOff": [404]
// }
type Server struct {
    http *http.Server

    FileEnable   bool // static file in public path enabled
    GzipEnable   bool // gzip output compress enabled
    GzipMinBytes int  // minimum bytes for gzip output

    statsInterval time.Duration // interval for output server stats
    errorLogOff   map[int]bool  // close error log for specific code

    totalReq uint64 // total requests since server start
    numReq   uint64 // num requests since last stats output
}

func (s *Server) Construct() {
    s.http = &http.Server{
        Addr:           DefaultServerAddr,
        ReadTimeout:    DefaultTimeout,
        WriteTimeout:   DefaultTimeout,
        MaxHeaderBytes: DefaultHeaderBytes,
        Handler:        s,
    }

    s.FileEnable = true
    s.GzipEnable = true
    s.GzipMinBytes = 1024

    s.statsInterval = 60 * time.Second
}

func (s *Server) SetAddr(addr string) {
    s.http.Addr = addr
}

func (s *Server) SetReadTimeout(timeout string) {
    s.http.ReadTimeout, _ = time.ParseDuration(timeout)
}

func (s *Server) SetWriteTimeout(timeout string) {
    s.http.WriteTimeout, _ = time.ParseDuration(timeout)
}

func (s *Server) SetMaxHeaderBytes(maxBytes int) {
    s.http.MaxHeaderBytes = maxBytes
}

func (s *Server) SetErrorLogOff(codes []interface{}) {
    s.errorLogOff = make(map[int]bool)
    for _, v := range codes {
        s.errorLogOff[Util.ToInt(v)] = true
    }
}

func (s *Server) SetStatsInterval(interval string) {
    s.statsInterval, _ = time.ParseDuration(interval)
}

func (s *Server) IsErrorLogOff(status int) bool {
    return s.errorLogOff[status]
}

func (s *Server) GetHttp() *http.Server {
    return s.http
}

func (s *Server) Serve() {
    enableDebugServer := App.GetConfig().GetBool("params.debugServer.enable", false)
    defer func() {
        if v := recover(); v != nil {
            GLogger.Fatal("%s trace[%s]", Util.ToString(v), Util.PanicTrace(TraceMaxDepth, false))
        }
        if enableDebugServer == true {
            dAddr := App.GetConfig().GetString("params.debugServer.addr", "0.0.0.0:8100")
            GLogger.Info("stop running debug http at %s", dAddr)
        }
        if App.GetMode() == ModeCmd {
            GLogger.Info("stop running command %s", flag.Lookup("cmd").Value)
        } else {
            GLogger.Info("stop running http at %s", s.http.Addr)
        }

        App.GetLog().Flush()
    }()
    // debug pprof
    if enableDebugServer == true {
        go func() {
            dAddr := App.GetConfig().GetString("params.debugServer.addr", "0.0.0.0:8100")
            GLogger.Info("start running debug http at %s", dAddr)
            ds := &http.Server{
                Addr:         dAddr,
                ReadTimeout:  40 * time.Second,
                WriteTimeout: 40 * time.Second,
            }
            ds.ListenAndServe()
        }()
    }
    if App.GetMode() == ModeCmd {
        GLogger.Info("start running command %s", flag.Lookup("cmd").Value)
        s.ServeCMD()
    } else {
        GLogger.Info("start running http at %s", s.http.Addr)
        wg := sync.WaitGroup{}
        wg.Add(1)

        // new goroutine to handle signal and statistics
        go s.handleSigAndStats(&wg)

        if e := s.http.ListenAndServe(); e != http.ErrServerClosed {
            GLogger.Fatal("ListenAndServe failed, %s", e)
        } else {
            wg.Wait() // wait completion of shutdown
        }
    }
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    atomic.AddUint64(&s.numReq, 1)

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

func (s *Server) ServeCMD() {
    ctx := &Context{}
    ctx.Init()

    s.handleRequest(ctx)
}

// goroutine to handle signal and statistics
func (s *Server) handleSigAndStats(wg *sync.WaitGroup) {
    sig := make(chan os.Signal)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    timer := time.Tick(s.statsInterval)

    for {
        select {
        case <-sig:
            ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
            s.http.Shutdown(ctx)
            goto end
        case <-timer:
            memStats := runtime.MemStats{}
            runtime.ReadMemStats(&memStats)

            numReq := atomic.SwapUint64(&s.numReq, 0)
            s.totalReq += numReq

            GLogger.Info("app stats, totalReq:%d, lastReq:%d, numGO:%d, sysMem(mb):%d, totalGC(ms):%d, lastGC(ms):%d",
                s.totalReq, numReq,
                runtime.NumGoroutine(),
                memStats.Sys/(1<<20),
                memStats.PauseTotalNs/1e6,
                memStats.PauseNs[(memStats.NumGC+255)%256]/1e6,
            )
        }
    }

end:
    wg.Done()
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
