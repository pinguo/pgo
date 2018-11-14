package pgo

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "reflect"

    "github.com/pinguo/pgo/Util"
)

// Controller the base class of web and cmd controller
type Controller struct {
    Object
}

// GetBindInfo get action map as extra binding info
func (c *Controller) GetBindInfo(v interface{}) interface{} {
    if _, ok := v.(IController); !ok {
        panic("param require a controller")
    }

    rt := reflect.ValueOf(v).Type()
    num := rt.NumMethod()
    actions := make(map[string]int)

    for i := 0; i < num; i++ {
        name := rt.Method(i).Name
        if len(name) > ActionLength && name[:ActionLength] == ActionPrefix {
            actions[name[ActionLength:]] = i
        }
    }

    return actions
}

// BeforeAction before action hook
func (c *Controller) BeforeAction(action string) {
}

// AfterAction after action hook
func (c *Controller) AfterAction(action string) {
}

// HandlePanic process unhandled action panic
func (c *Controller) HandlePanic(v interface{}) {
    status := http.StatusInternalServerError
    switch e := v.(type) {
    case *Exception:
        status = e.GetStatus()
        c.OutputJson(EmptyObject, status, e.GetMessage())
    default:
        c.OutputJson(EmptyObject, status)
    }

    c.GetContext().Error("%s, trace[%s]", Util.ToString(v), Util.PanicTrace(TraceMaxDepth, false))
}

// Redirect output redirect response
func (c *Controller) Redirect(location string, permanent bool) {
    ctx := c.GetContext()
    ctx.SetHeader("Location", location)
    if permanent {
        ctx.End(http.StatusMovedPermanently, nil)
    } else {
        ctx.End(http.StatusFound, nil)
    }
}

// OutputJson output json response
func (c *Controller) OutputJson(data interface{}, status int, msg ...string) {
    ctx := c.GetContext()
    message := App.GetStatus().GetText(status, ctx, msg...)
    output, e := json.Marshal(map[string]interface{}{
        "status":  status,
        "message": message,
        "data":    data,
    })

    if e != nil {
        panic(fmt.Sprintf("failed to marshal json, %s", e))
    }

    ctx.PushLog("status", status)
    ctx.SetHeader("Content-Type", "application/json; charset=utf-8")
    ctx.End(http.StatusOK, output)
}

// OutputJsonp output jsonp response
func (c *Controller) OutputJsonp(callback string, data interface{}, status int, msg ...string) {
    ctx := c.GetContext()
    message := App.GetStatus().GetText(status, ctx, msg...)
    output, e := json.Marshal(map[string]interface{}{
        "status":  status,
        "message": message,
        "data":    data,
    })

    if e != nil {
        panic(fmt.Sprintf("failed to marshal json, %s", e))
    }

    buf := &bytes.Buffer{}
    buf.WriteString(callback + "(")
    buf.Write(output)
    buf.WriteString(")")

    ctx.PushLog("status", status)
    ctx.SetHeader("Content-Type", "text/javascript; charset=utf-8")
    ctx.End(http.StatusOK, buf.Bytes())
}

// OutputView output rendered view
func (c *Controller) OutputView(view string, data interface{}, contentType ...string) {
    ctx := c.GetContext()
    contentType = append(contentType, "text/html; charset=utf-8")
    ctx.PushLog("status", http.StatusOK)
    ctx.SetHeader("Content-Type", contentType[0])
    ctx.End(http.StatusOK, App.GetView().Render(view, data))
}
