package pgo

import (
    "flag"
    "fmt"
    "net/http"
    "os"
    "reflect"
    "strings"
    "time"

    "github.com/pinguo/pgo/Util"
)

type objectItem struct {
    name string
    rv   reflect.Value
}

// Context pgo request context, context is not goroutine
// safe, copy context to use in other goroutines
type Context struct {
    server   *Server
    response Response
    input    *http.Request
    output   http.ResponseWriter

    startTime    time.Time
    controllerId string
    actionId     string
    userData     map[string]interface{}
    plugins      []IPlugin
    index        int
    objects      []objectItem

    Profiler
    Logger
}

// start plugin chain process
func (c *Context) process(plugins []IPlugin) {
    // generate request id
    logId := c.GetHeader("X-Log-Id", "")
    if logId == "" {
        logId = Util.GenUniqueId()
    }

    // reset properties
    c.startTime = time.Now()
    c.controllerId = ""
    c.actionId = ""
    c.userData = nil
    c.plugins = plugins
    c.index = -1
    c.Profiler.reset()
    c.Logger.init(App.GetName(), logId, App.GetLog())

    // finish response
    defer c.finish()

    // process request
    c.Next()

}

// finish process request
func (c *Context) finish() {
    // process unhandled panic
    if v := recover(); v != nil {
        status := http.StatusInternalServerError
        switch e := v.(type) {
        case *Exception:
            status = e.GetStatus()
            c.End(status, []byte(App.GetStatus().GetText(status, c, e.GetMessage())))
        default:
            c.End(status, []byte(http.StatusText(status)))
        }

        c.Error("%s, trace[%s]", Util.ToString(v), Util.PanicTrace(TraceMaxDepth, false))
    }

    // write header if not yet
    c.response.finish()

    // write access log
    if c.server.enableAccessLog {
        c.Notice("%s %s %d %d %dms pushlog[%s] profile[%s] counting[%s]",
            c.GetMethod(), c.GetPath(), c.GetStatus(), c.GetSize(), c.GetElapseMs(),
            c.GetPushLogString(), c.GetProfileString(), c.GetCountingString())
    }

    // clean objects
    c.clean()
}

// cache object in context
func (c *Context) cache(name string, rv reflect.Value) {
    if App.GetMode() == ModeWeb && len(c.objects) < MaxCacheObjects {
        c.objects = append(c.objects, objectItem{name, rv})
    }
}

// clean all cached objects
func (c *Context) clean() {
    container, num := App.GetContainer(), len(c.objects)
    for i := 0; i < num; i++ {
        name, rv := c.objects[i].name, c.objects[i].rv
        container.Put(name, rv)
    }

    // reset object pool to empty
    if num > 0 {
        c.objects = c.objects[:0]
    }
}

// Next start running plugin chain
func (c *Context) Next() {
    c.index++
    for num := len(c.plugins); c.index < num; c.index++ {
        c.plugins[c.index].HandleRequest(c)
    }
}

// Abort stop running plugin chain
func (c *Context) Abort() {
    c.index = MaxPlugins
}

// Copy copy context
func (c *Context) Copy() *Context {
    cp := *c
    cp.Profiler.reset()
    cp.userData = nil
    cp.plugins = nil
    cp.index = MaxPlugins
    cp.objects = nil
    return &cp
}

// GetElapseMs get elapsed ms since request start
func (c *Context) GetElapseMs() int {
    elapse := time.Now().Sub(c.startTime)
    return int(elapse.Nanoseconds() / 1e6)
}

// GetLogId get log id of current context
func (c *Context) GetLogId() string {
    return c.Logger.logId
}

// GetStatus get response status
func (c *Context) GetStatus() int {
    return c.response.status
}

// GetSize get response size
func (c *Context) GetSize() int {
    return c.response.size
}

// SetInput
func (c *Context) SetInput(r *http.Request) {
    c.input = r
}

// SetInput
func (c *Context) GetInput() *http.Request {
    return c.input
}

// SetOutput
func (c *Context) SetOutput(w http.ResponseWriter) {
    c.output = w
}

// GetOutput
func (c *Context) GetOutput() http.ResponseWriter {
    return c.output
}

// SetControllerId
func (c *Context) SetControllerId(id string) {
    c.controllerId = id
}

// GetControllerId
func (c *Context) GetControllerId() string {
    return c.controllerId
}

// SetActionId
func (c *Context) SetActionId(id string) {
    c.actionId = id
}

// GetActionId
func (c *Context) GetActionId() string {
    return c.actionId
}

// SetUserData set user data to current context
func (c *Context) SetUserData(key string, data interface{}) {
    if nil == c.userData {
        c.userData = make(map[string]interface{})
    }

    c.userData[key] = data
}

// GetUserData get user data from current context
func (c *Context) GetUserData(key string, dft interface{}) interface{} {
    if data, ok := c.userData[key]; ok {
        return data
    }

    return dft
}

// GetMethod get request method
func (c *Context) GetMethod() string {
    if c.input != nil {
        return c.input.Method
    }

    return "CMD"
}

// GetQuery get first url query value by name
func (c *Context) GetQuery(name, dft string) string {
    if c.input != nil {
        v := c.input.URL.Query().Get(name)
        if len(v) > 0 {
            return v
        }
    }

    return dft
}

// GetQueryAll get first value of all url queries
func (c *Context) GetQueryAll() map[string]string {
    m := make(map[string]string)
    if c.input != nil {
        queries := c.input.URL.Query()
        for k, v := range queries {
            if len(v) > 0 {
                m[k] = v[0]
            } else {
                m[k] = ""
            }
        }
    }

    return m
}

// GetPost get first post value by name
func (c *Context) GetPost(name, dft string) string {
    if c.input != nil {
        v := c.input.PostFormValue(name)
        if len(v) > 0 {
            return v
        }
    }

    return dft
}

// GetPostAll get first value of all posts
func (c *Context) GetPostAll() map[string]string {
    m := make(map[string]string)
    if c.input != nil {
        // make sure c.input.ParseMultipartForm has been called
        c.input.PostFormValue("")
        for k, v := range c.input.PostForm {
            if len(v) > 0 {
                m[k] = v[0]
            } else {
                m[k] = ""
            }
        }
    }

    return m
}

// GetParam get first param value by name, post take precedence over get
func (c *Context) GetParam(name, dft string) string {
    if c.input != nil {
        v := c.input.FormValue(name)
        if len(v) > 0 {
            return v
        }
    }

    return dft
}

// GetParamAll get first value of all params, post take precedence over get
func (c *Context) GetParamAll() map[string]string {
    m := make(map[string]string)
    if c.input != nil {
        // make sure c.input.ParseMultipartForm has been called
        c.input.FormValue("")
        for k, v := range c.input.Form {
            if len(v) > 0 {
                m[k] = v[0]
            } else {
                m[k] = ""
            }
        }
    }

    return m
}

// GetParamMap get map value from GET/POST
func (c *Context) GetParamMap(name string) map[string]string {
    // TODO name[k1]=v1&name[k2]=v2
    return nil
}

// GetQueryMap get map value from GET
func (c *Context) GetQueryMap(name string) map[string]string {
    // TODO name[k1]=v1&name[k2]=v2
    return nil
}

// GetPostMap get map value from POST
func (c *Context) GetPostMap(name string) map[string]string {
    // TODO name[k1]=v1&name[k2]=v2
    return nil
}

// GetParamArray get array value from GET/POST
func (c *Context) GetParamArray(name string) []string {
    // TODO name[]=v1&name[]=v2
    return nil
}

// GetQueryArray get array value from GET
func (c *Context) GetQueryArray(name string) []string {
    // TODO name[]=v1&name[]=v2
    return nil
}

// GetPostArray get array value from POST
func (c *Context) GetPostArray(name string) []string {
    // TODO name[]=v1&name[]=v2
    return nil
}

// GetCookie get first cookie value by name
func (c *Context) GetCookie(name, dft string) string {
    if c.input != nil {
        v, e := c.input.Cookie(name)
        if e == nil && len(v.Value) > 0 {
            return v.Value
        }
    }

    return dft
}

// GetCookieAll get first value of all cookies
func (c *Context) GetCookieAll() map[string]string {
    m := make(map[string]string)
    if c.input != nil {
        cookies := c.input.Cookies()
        for _, cookie := range cookies {
            if _, ok := m[cookie.Name]; !ok {
                m[cookie.Name] = cookie.Value
            }
        }
    }

    return m
}

// GetHeader get first header value by name，name is case-insensitive
func (c *Context) GetHeader(name, dft string) string {
    if c.input != nil {
        v := c.input.Header.Get(name)
        if len(v) > 0 {
            return v
        }
    }

    return dft
}

// GetHeaderAll get first value of all headers
func (c *Context) GetHeaderAll() map[string]string {
    m := make(map[string]string)
    if c.input != nil {
        for k, v := range c.input.Header {
            if len(v) > 0 {
                m[k] = v[0]
            } else {
                m[k] = ""
            }
        }
    }

    return m
}

// GetPath get request path
func (c *Context) GetPath() string {
    // for web
    if c.input != nil {
        return c.input.URL.Path
    }

    // for cmd
    f := flag.Lookup("cmd")
    if f != nil && len(f.Value.String()) > 0 {
        return f.Value.String()
    }

    return "/"
}

// GetClientIp get client ip
func (c *Context) GetClientIp() string {
    if xff := c.GetHeader("X-Forwarded-For", ""); len(xff) > 0 {
        if pos := strings.IndexByte(xff, ','); pos > 0 {
            return strings.TrimSpace(xff[:pos])
        } else {
            return xff
        }
    }

    if ip := c.GetHeader("X-Client-Ip", ""); len(ip) > 0 {
        return ip
    }

    if ip := c.GetHeader("X-Real-Ip", ""); len(ip) > 0 {
        return ip
    }

    if c.input != nil && len(c.input.RemoteAddr) > 0 {
        pos := strings.LastIndexByte(c.input.RemoteAddr, ':')
        if pos > 0 {
            return c.input.RemoteAddr[:pos]
        } else {
            return c.input.RemoteAddr
        }
    }

    return ""
}

// validate query param, return string validator
func (c *Context) ValidateQuery(name string, dft ...interface{}) *StringValidator {
    return ValidateString(c.GetQuery(name, ""), name, dft...)
}

// validate post param, return string validator
func (c *Context) ValidatePost(name string, dft ...interface{}) *StringValidator {
    return ValidateString(c.GetPost(name, ""), name, dft...)
}

// validate get/post param, return string validator
func (c *Context) ValidateParam(name string, dft ...interface{}) *StringValidator {
    return ValidateString(c.GetParam(name, ""), name, dft...)
}

// set response header, no effect if any header has sent
func (c *Context) SetHeader(name, value string) {
    if c.output != nil {
        c.output.Header().Set(name, value)
    }
}

// convenient way to set response cookie
func (c *Context) SetCookie(cookie *http.Cookie) {
    if c.output != nil {
        http.SetCookie(c.output, cookie)
    }
}

// send response
func (c *Context) End(status int, data []byte) {
    if c.output != nil {
        c.SetHeader("X-Log-Id", c.Logger.logId)
        c.SetHeader("X-Cost-Time", fmt.Sprintf("%dms", c.GetElapseMs()))

        c.output.WriteHeader(status)
        c.output.Write(data)
    } else if len(data) > 0 {
        os.Stdout.Write(data)
        os.Stdout.WriteString("\n")
    }
}
