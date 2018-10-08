package Memcache

import (
    "pgo"
    "sync"
    "sync/atomic"
    "time"
)

// Memcache Client component, configuration:
// {
//     "class": "@pgo/Client/Memcache/Client",
//     "prefix": "pgo_",
//     "maxIdleConn": 10,
//     "maxIdleTime": "60s",
//     "netTimeout": "1s",
//     "probInterval": "0s",
//     "servers": [
//         "127.0.0.1:11211",
//         "127.0.0.1:11212"
//     ]
// }
type Client struct {
    Pool
}

func (c *Client) Get(key string) *pgo.Value {
    if item := c.Retrieve(CmdGet, key); item != nil {
        return pgo.NewValue(item.Data)
    }
    return pgo.NewValue(nil)
}

func (c *Client) MGet(keys []string) map[string]*pgo.Value {
    result := make(map[string]*pgo.Value)
    for _, key := range keys {
        result[key] = pgo.NewValue(nil)
    }

    if items := c.MultiRetrieve(CmdGet, keys); len(items) > 0 {
        for _, item := range items {
            result[item.Key] = pgo.NewValue(item.Data)
        }
    }
    return result
}

func (c *Client) Set(key string, value interface{}, expire ...time.Duration) bool {
    return c.Store(CmdSet, &Item{Key: key, Data: pgo.Encode(value)}, expire...)
}

func (c *Client) MSet(items map[string]interface{}, expire ...time.Duration) bool {
    newItems := make([]*Item, 0, len(items))
    for key, value := range items {
        newItems = append(newItems, &Item{Key: key, Data: pgo.Encode(value)})
    }
    return c.MultiStore(CmdSet, newItems, expire...)
}

func (c *Client) Add(key string, value interface{}, expire ...time.Duration) bool {
    return c.Store(CmdAdd, &Item{Key: key, Data: pgo.Encode(value)}, expire...)
}

func (c *Client) MAdd(items map[string]interface{}, expire ...time.Duration) bool {
    newItems := make([]*Item, 0, len(items))
    for key, value := range items {
        newItems = append(newItems, &Item{Key: key, Data: pgo.Encode(value)})
    }
    return c.MultiStore(CmdAdd, newItems, expire...)
}

func (c *Client) Del(key string) bool {
    newKey := c.BuildKey(key)
    conn := c.GetConnByKey(newKey)
    defer conn.Close(false)

    return conn.Delete(newKey)
}

func (c *Client) MDel(keys []string) bool {
    addrKeys, _ := c.AddrNewKeys(keys)
    wg, success := new(sync.WaitGroup), uint32(0)

    wg.Add(len(addrKeys))
    for addr, keys := range addrKeys {
        go c.RunAddrFunc(addr, keys, wg, func(conn *Conn, keys []string) {
            for _, key := range keys {
                // extend deadline for every operation
                conn.ExtendDeadLine()
                if ok := conn.Delete(key); ok {
                    atomic.AddUint32(&success, 1)
                }
            }
        })
    }

    wg.Wait()
    return success == uint32(len(keys))
}

func (c *Client) Exists(key string) bool {
    return c.Get(key) != nil
}

func (c *Client) Incr(key string, delta int) int {
    newKey := c.BuildKey(key)
    conn := c.GetConnByKey(newKey)
    defer conn.Close(false)

    return conn.Increment(newKey, delta)
}

func (c *Client) Retrieve(cmd, key string) *Item {
    newKey := c.BuildKey(key)
    conn := c.GetConnByKey(newKey)
    defer conn.Close(false)

    if items := conn.Retrieve(cmd, newKey); len(items) == 1 {
        return items[0]
    }
    return nil
}

func (c *Client) MultiRetrieve(cmd string, keys []string) []*Item {
    result := make([]*Item, 0, len(keys))
    addrKeys, newKeys := c.AddrNewKeys(keys)
    lock, wg := new(sync.Mutex), new(sync.WaitGroup)

    wg.Add(len(addrKeys))
    for addr, keys := range addrKeys {
        go c.RunAddrFunc(addr, keys, wg, func(conn *Conn, keys []string) {
            if items := conn.Retrieve(cmd, keys...); len(items) > 0 {
                lock.Lock()
                defer lock.Unlock()
                for _, item := range items {
                    item.Key = newKeys[item.Key]
                    result = append(result, item)
                }
            }
        })
    }

    wg.Wait()
    return result
}

func (c *Client) Store(cmd string, item *Item, expire ...time.Duration) bool {
    item.Key = c.BuildKey(item.Key)
    conn := c.GetConnByKey(item.Key)
    defer conn.Close(false)

    expire = append(expire, defaultExpire)
    return conn.Store(cmd, item, int(expire[0]/time.Second))
}

func (c *Client) MultiStore(cmd string, items []*Item, expire ...time.Duration) bool {
    expire = append(expire, defaultExpire)
    addrItems := make(map[string][]*Item)
    wg, success := new(sync.WaitGroup), uint32(0)

    for _, item := range items {
        item.Key = c.BuildKey(item.Key)
        addr := c.GetAddrByKey(item.Key)
        addrItems[addr] = append(addrItems[addr], item)
    }

    wg.Add(len(addrItems))
    for addr := range addrItems {
        go c.RunAddrFunc(addr, nil, wg, func(conn *Conn, keys []string) {
            for _, item := range addrItems[addr] {
                conn.ExtendDeadLine() // extend deadline for every store
                if ok := conn.Store(cmd, item, int(expire[0]/time.Second)); ok {
                    atomic.AddUint32(&success, 1)
                }
            }
        })
    }

    wg.Wait()
    return success == uint32(len(items))
}
