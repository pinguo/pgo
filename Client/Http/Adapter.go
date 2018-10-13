package Http

import (
    "net/http"
    "sync"
    "time"

    "github.com/pinguo/pgo"
    "github.com/pinguo/pgo/Util"
)

// Adapter of Http Client, add context support.
// usage: http := this.GetObject("@pgo/Client/Http/Adapter").(*Adapter)
type Adapter struct {
    pgo.Object
    client       *Client
    panicRecover bool
}

func (a *Adapter) Construct(componentId ...string) {
    id := defaultComponentId
    if len(componentId) > 0 {
        id = componentId[0]
    }

    a.client = pgo.App.Get(id).(*Client)
    a.panicRecover = true
}

func (a *Adapter) SetPanicRecover(v bool) {
    a.panicRecover = v
}

func (a *Adapter) GetClient() *Client {
    return a.client
}

func (a *Adapter) handlePanic() {
    if a.panicRecover {
        if v := recover(); v != nil {
            a.GetContext().Error(Util.ToString(v))
        }
    }
}

// Get perform a get request
func (a *Adapter) Get(addr string, data interface{}, option ...*Option) *http.Response {
    profile := baseUrl(addr)
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Get(addr, data, option...)
}

// Post perform a post request
func (a *Adapter) Post(addr string, data interface{}, option ...*Option) *http.Response {
    profile := baseUrl(addr)
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Post(addr, data, option...)
}

// Do perform a single request
func (a *Adapter) Do(req *http.Request, option ...*Option) *http.Response {
    profile := baseUrl(req.URL.String())
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Do(req, option...)
}

// DoMulti perform multi requests concurrently
func (a *Adapter) DoMulti(reqArr []*http.Request, option ...*Option) []*http.Response {
    if optNum := len(option); optNum != 0 && optNum != len(reqArr) {
        panic("http multi request invalid num of options")
    }

    profile := "Http.DoMulti"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    lock, wg := new(sync.Mutex), new(sync.WaitGroup)
    resArr := make([]*http.Response, len(reqArr))

    fn := func(k int) {
        defer func() {
            recover()
            wg.Done()
        }()

        start, profile := time.Now(), baseUrl(reqArr[k].URL.String())
        var res *http.Response

        defer func() {
            lock.Lock()
            a.GetContext().ProfileAdd(profile, time.Since(start)/1e6)
            resArr[k] = res
            lock.Unlock()
        }()

        if len(option) > 0 && option[k] != nil {
            res = a.client.Do(reqArr[k], option[k])
        } else {
            res = a.client.Do(reqArr[k])
        }
    }

    wg.Add(len(reqArr))
    for k := range reqArr {
        go fn(k)
    }

    wg.Wait()
    return resArr
}
