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
func (a *Adapter) DoMulti(requests []*http.Request, option ...*Option) []*http.Response {
    if num := len(option); num != 0 && num != len(requests) {
        panic("http multi request invalid num of options")
    }

    profile := "Http.DoMulti"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    lock, wg := new(sync.Mutex), new(sync.WaitGroup)
    responses := make([]*http.Response, len(requests))

    fn := func(k int) {
        start, profile := time.Now(), baseUrl(requests[k].URL.String())
        var res *http.Response

        defer func() {
            if v := recover(); v != nil {
                a.GetContext().Error(Util.ToString(v))
            }

            lock.Lock()
            a.GetContext().ProfileAdd(profile, time.Since(start)/1e6)
            responses[k] = res
            lock.Unlock()
            wg.Done()
        }()

        if len(option) > 0 && option[k] != nil {
            res = a.client.Do(requests[k], option[k])
        } else {
            res = a.client.Do(requests[k])
        }
    }

    wg.Add(len(requests))
    for k := range requests {
        go fn(k)
    }

    wg.Wait()
    return responses
}
