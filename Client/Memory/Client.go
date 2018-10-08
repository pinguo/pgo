package Memory

import (
    "fmt"
    "sync"
    "time"

    "github.com/pinguo/pgo"
    "github.com/pinguo/pgo/Util"
)

type item struct {
    value  interface{}
    expire time.Time
}

func (i item) isExpired() bool {
    return !i.expire.IsZero() && time.Since(i.expire) > 0
}

// Memory Client component, configuration:
// {
//     "class": "@pgo/Client/Memory/Client",
//     "gcInterval": "60s",
//     "gcMaxItems": 1000
// }
type Client struct {
    lock       sync.RWMutex
    items      map[string]*item
    gcInterval time.Duration
    gcMaxItems int
}

func (c *Client) Construct() {
    c.items = make(map[string]*item)
    c.gcInterval = defaultGcInterval
    c.gcMaxItems = defaultGcMaxItems
}

func (c *Client) Init() {
    go c.gcLoop()
}

func (c *Client) SetGcInterval(v string) {
    if gcInterval, e := time.ParseDuration(v); e != nil {
        panic(fmt.Sprintf(errSetProp, "gcInterval", e.Error()))
    } else {
        c.gcInterval = gcInterval
    }
}

func (c *Client) SetGcMaxItems(gcMaxItems int) {
    if gcMaxItems > 0 {
        c.gcMaxItems = gcMaxItems
    }
}

func (c *Client) Get(key string) *pgo.Value {
    c.lock.RLock()
    defer c.lock.RUnlock()

    if item := c.items[key]; item != nil && !item.isExpired() {
        return pgo.NewValue(item.value)
    }

    return pgo.NewValue(nil)
}

func (c *Client) MGet(keys []string) map[string]*pgo.Value {
    c.lock.RLock()
    defer c.lock.RUnlock()

    result := make(map[string]*pgo.Value)
    for _, key := range keys {
        if item := c.items[key]; item != nil && !item.isExpired() {
            result[key] = pgo.NewValue(item.value)
        } else {
            result[key] = pgo.NewValue(nil)
        }
    }

    return result
}

func (c *Client) Set(key string, value interface{}, expire ...time.Duration) bool {
    c.lock.Lock()
    defer c.lock.Unlock()

    expire, now := append(expire, defaultExpire), time.Now()
    c.items[key] = &item{
        value:  value,
        expire: now.Add(expire[0]),
    }

    return true
}

func (c *Client) MSet(items map[string]interface{}, expire ...time.Duration) bool {
    c.lock.Lock()
    defer c.lock.Unlock()

    expire, now := append(expire, defaultExpire), time.Now()
    for key, value := range items {
        c.items[key] = &item{
            value:  value,
            expire: now.Add(expire[0]),
        }
    }
    return true
}

func (c *Client) Add(key string, value interface{}, expire ...time.Duration) bool {
    c.lock.Lock()
    defer c.lock.Unlock()

    expire, now := append(expire, defaultExpire), time.Now()
    if old := c.items[key]; old == nil || old.isExpired() {
        c.items[key] = &item{
            value:  value,
            expire: now.Add(expire[0]),
        }
        return true
    }
    return false
}

func (c *Client) MAdd(items map[string]interface{}, expire ...time.Duration) bool {
    c.lock.Lock()
    defer c.lock.Unlock()

    expire, now, success := append(expire, defaultExpire), time.Now(), 0
    for key, value := range items {
        if old := c.items[key]; old == nil || old.isExpired() {
            c.items[key] = &item{
                value:  value,
                expire: now.Add(expire[0]),
            }
            success++
        }
    }

    return success == len(items)
}

func (c *Client) Del(key string) bool {
    c.lock.Lock()
    defer c.lock.Unlock()

    if _, ok := c.items[key]; ok {
        delete(c.items, key)
        return true
    }
    return false
}

func (c *Client) MDel(keys []string) bool {
    c.lock.Lock()
    defer c.lock.Unlock()

    success := 0
    for _, key := range keys {
        if _, ok := c.items[key]; ok {
            delete(c.items, key)
            success++
        }
    }

    return success == len(keys)
}

func (c *Client) Exists(key string) bool {
    c.lock.RLock()
    defer c.lock.RUnlock()

    _, ok := c.items[key]
    return ok
}

func (c *Client) Incr(key string, delta int) int {
    c.lock.Lock()
    defer c.lock.Unlock()

    cur := c.items[key]
    if cur == nil {
        cur = &item{value: 0}
        c.items[key] = cur
    }

    newVal := Util.ToInt(cur.value) + delta
    cur.value = newVal
    return newVal
}

func (c *Client) gcLoop() {
    if c.gcInterval < minGcInterval || c.gcInterval > maxGcInterval {
        c.gcInterval = defaultGcInterval
    }

    getExpireKeys := func() []string {
        c.lock.RLock()
        defer c.lock.RUnlock()

        keys, now := make([]string, 0), time.Now()
        for key, item := range c.items {
            if !item.expire.IsZero() && item.expire.Sub(now) < 0 {
                keys = append(keys, key)
                if len(keys) >= c.gcMaxItems {
                    break
                }
            }
        }
        return keys
    }

    clearExpiredKeys := func(keys []string) {
        c.lock.Lock()
        defer c.lock.Unlock()

        for _, key := range keys {
            delete(c.items, key)
        }
    }

    for {
        <-time.After(c.gcInterval)
        if expiredKeys := getExpireKeys(); len(expiredKeys) > 0 {
            clearExpiredKeys(expiredKeys)
        }
    }
}
