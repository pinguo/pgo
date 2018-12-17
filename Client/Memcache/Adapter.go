package Memcache

import (
    "time"

    "github.com/pinguo/pgo"
    "github.com/pinguo/pgo/Util"
)

// Adapter of Memcache Client, add context support.
// usage: mc := this.GetObject(Memcache.AdapterClass.(*Memcache.Adapter)
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
    profile := "Memcache.Get"
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
    profile := "Memcache.MGet"
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
    profile := "Memcache.Set"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Set(key, value, expire...)
}

func (a *Adapter) MSet(items map[string]interface{}, expire ...time.Duration) bool {
    profile := "Memcache.MSet"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.MSet(items, expire...)
}

func (a *Adapter) Add(key string, value interface{}, expire ...time.Duration) bool {
    profile := "Memcache.Add"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Add(key, value, expire...)
}

func (a *Adapter) MAdd(items map[string]interface{}, expire ...time.Duration) bool {
    profile := "Memcache.MAdd"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStart(profile)
    defer a.handlePanic()

    return a.client.MAdd(items, expire...)
}

func (a *Adapter) Del(key string) bool {
    profile := "Memcache.Del"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Del(key)
}

func (a *Adapter) MDel(keys []string) bool {
    profile := "Memcache.MDel"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.MDel(keys)
}

func (a *Adapter) Exists(key string) bool {
    profile := "Memcache.Exists"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Exists(key)
}

func (a *Adapter) Incr(key string, delta int) int {
    profile := "Memcache.Incr"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Incr(key, delta)
}

func (a *Adapter) Retrieve(cmd, key string) *Item {
    profile := "Memcache.Retrieve"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Retrieve(cmd, key)
}

func (a *Adapter) MultiRetrieve(cmd string, keys []string) []*Item {
    profile := "Memcache.MultiRetrieve"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.MultiRetrieve(cmd, keys)
}

func (a *Adapter) Store(cmd string, item *Item, expire ...time.Duration) bool {
    profile := "Memcache.Store"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.Store(cmd, item, expire...)
}

func (a *Adapter) MultiStore(cmd string, items []*Item, expire ...time.Duration) bool {
    profile := "Memcache.MultiStore"
    a.GetContext().ProfileStart(profile)
    defer a.GetContext().ProfileStop(profile)
    defer a.handlePanic()

    return a.client.MultiStore(cmd, items, expire...)
}
