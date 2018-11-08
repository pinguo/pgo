package pgo

import (
    "bytes"
    "compress/gzip"
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/pinguo/pgo/Util"
)

// Context pgo request context.
type Context struct {
    input        *http.Request
    output       http.ResponseWriter
    startTime    time.Time
    logId        string
    controllerId string
    actionId     string
    userData     map[string]interface{}
    *Profiler
    *Logger
}

func (c *Context) Init() {
    c.startTime = time.Now()
    c.Profiler = App.GetLog().GetProfiler()

    if App.GetMode() == ModeCmd {
        c.Logger = GLogger()
    } else {
        c.Logger = App.GetLog().GetLogger(App.name, c.GetLogId())
    }
}

func (c *Context) SetInput(r *http.Request) {
    c.input = r
}

func (c *Context) GetInput() *http.Request {
    return c.input
}

func (c *Context) SetOutput(w http.ResponseWriter) {
    c.output = w
}

func (c *Context) GetOutput() http.ResponseWriter {
    return c.output
}

func (c *Context) GetElapseMs() int {
    elapse := time.Now().Sub(c.startTime)
    return int(elapse.Nanoseconds() / 1e6)
}

func (c *Context) GetLogId() string {
    if len(c.logId) == 0 {
        c.logId = c.GetHeader("X-Log-Id", "")
        if len(c.logId) == 0 {
            c.logId = Util.GenUniqueId()
        }
    }

    return c.logId
}

func (c *Context) SetControllerId(id string) {
    c.controllerId = id
}

func (c *Context) GetControllerId() string {
    return c.controllerId
}

func (c *Context) SetActionId(id string) {
    c.actionId = id
}

func (c *Context) GetActionId() string {
    return c.actionId
}

func (c *Context) SetUserData(key string, data interface{}) {
    if nil == c.userData {
        c.userData = make(map[string]interface{})
    }

    c.userData[key] = data
}

func (c *Context) GetUserData(key string, dft interface{}) interface{} {
    if nil != c.userData {
        if data, ok := c.userData[key]; ok {
            return data
        }
    }

    return dft
}

// get request method
func (c *Context) GetMethod() string {
    if c.input != nil {
        return c.input.Method
    }

    return ""
}

// get first url query value by name
func (c *Context) GetQuery(name, dft string) string {
    if c.input != nil {
        v := c.input.URL.Query().Get(name)
        if len(v) > 0 {
            return v
        }
    }

    return dft
}

// get first value of all url queries
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

// get first post value by name
func (c *Context) GetPost(name, dft string) string {
    if c.input != nil {
        v := c.input.PostFormValue(name)
        if len(v) > 0 {
            return v
        }
    }

    return dft
}

// get first value of all posts
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

// get first param value by name, post take precedence over get
func (c *Context) GetParam(name, dft string) string {
    if c.input != nil {
        v := c.input.FormValue(name)
        if len(v) > 0 {
            return v
        }
    }

    return dft
}

// get first value of all params, post take precedence over get
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

// get params parsed from a complex url, like parse_str() in php
func (c *Context) GetFixedParamMap() map[string]interface{} {
    // TODO k1=12&k2[0]=aa&k2[1]=bb&k2[2]=cc&k3[m1]=v1&k3[m2]=v2&k3[m3][0]=10
    return nil
}

// get first cookie value by name
func (c *Context) GetCookie(name, dft string) string {
    if c.input != nil {
        v, e := c.input.Cookie(name)
        if e == nil && len(v.Value) > 0 {
            return v.Value
        }
    }

    return dft
}

// get first value of all cookies
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

// get first header value by name，name is case-insensitive
func (c *Context) GetHeader(name, dft string) string {
    if c.input != nil {
        v := c.input.Header.Get(name)
        if len(v) > 0 {
            return v
        }
    }

    return dft
}

// get first value of all headers
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

// get request path
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

// get client ip
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

// get json decoded body
func (c *Context) GetJsonBody(target interface{}) error {
    ct := c.GetHeader("Content-Type", "")
    if !strings.HasPrefix(ct, "application/json") {
        return errors.New("invalid content-type: " + ct)
    }

    return json.NewDecoder(c.input.Body).Decode(target)
}

// get raw body bytes
func (c *Context) GetRawBody() []byte {
    if c.input == nil {
        return nil
    }

    buf := &bytes.Buffer{}
    buf.ReadFrom(c.input.Body)
    return buf.Bytes()
}

// validate query param, return string validator
func (c *Context) ValidateQuery(name string, dft ...interface{}) *StringValidator {
    return ValidateString(c.GetQuery(name, ""), name, dft...)
}

// validate post param, return string validator
func (c *Context) ValidatePost(name string, dft ...interface{}) *StringValidator {
    return ValidateString(c.GetPost(name, ""), name, dft...)
}

// validate get or post param, return string validator
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

// send http response, gzip data if possible
func (c *Context) End(status int, data []byte) {
    if c.output != nil {
        if len(http.StatusText(status)) == 0 {
            status = http.StatusOK
        }

        c.SetHeader("X-Log-Id", c.GetLogId())
        c.SetHeader("X-Cost-Time", fmt.Sprintf("%dms", c.GetElapseMs()))

        svr := App.GetServer()
        if svr.GzipEnable && len(data) > svr.GzipMinBytes {
            ae := c.GetHeader("Accept-Encoding", "")
            if pos := strings.Index(ae, "gzip"); pos != -1 {
                c.SetHeader("Content-Encoding", "gzip")
                c.output.WriteHeader(status)
                gz := gzip.NewWriter(c.output)
                gz.Write(data)
                gz.Close()
                return
            }
        }

        c.output.WriteHeader(status)
        c.output.Write(data)
    } else if len(data) > 0 {
        os.Stdout.Write(data)
    }
}
