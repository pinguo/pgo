package Memory

import (
    "time"

    "github.com/pinguo/pgo"
    "github.com/pinguo/pgo/Util"
)

// Adapter of Memory Client, add context support.
// usage: memory := this.GetObject("@pgo/Client/Memory/Adapter").(*Adapter)
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

func (a *Adapter) Get(key string) *pgo.Value {
    profile := "Memory.Get"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    res, hit := a.client.Get(key), 0
    if res != nil && res.Valid() {
        hit = 1
    }

    a.GetContext().Counting(profile, hit, 1)
    return res
}

func (a *Adapter) MGet(keys []string) map[string]*pgo.Value {
    profile := "Memory.MGet"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    res, hit := a.client.MGet(keys), 0
    for _, v := range res {
        if v != nil && v.Valid() {
            hit += 1
        }
    }

    a.GetContext().Counting(profile, hit, len(keys))
    return res
}

func (a *Adapter) Set(key string, value interface{}, expire ...time.Duration) bool {
    profile := "Memory.Set"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Set(key, value, expire...)
}

func (a *Adapter) MSet(items map[string]interface{}, expire ...time.Duration) bool {
    profile := "Memory.MSet"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.MSet(items, expire...)
}

func (a *Adapter) Add(key string, value interface{}, expire ...time.Duration) bool {
    profile := "Memory.Add"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Add(key, value, expire...)
}

func (a *Adapter) MAdd(items map[string]interface{}, expire ...time.Duration) bool {
    profile := "Memory.MAdd"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStart(profile)
    defer a.handlePanic()

    return a.client.MAdd(items, expire...)
}

func (a *Adapter) Del(key string) bool {
    profile := "Memory.Del"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Del(key)
}

func (a *Adapter) MDel(keys []string) bool {
    profile := "Memory.MDel"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.MDel(keys)
}

func (a *Adapter) Exists(key string) bool {
    profile := "Memory.Exists"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Exists(key)
}

func (a *Adapter) Incr(key string, delta int) int {
    profile := "Memory.Incr"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Incr(key, delta)
}
