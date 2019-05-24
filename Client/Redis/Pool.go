package Redis

import (
    "fmt"
    "net"
    "strings"
    "sync"
    "time"

    "github.com/pinguo/pgo"
    "github.com/pinguo/pgo/Util"
)

type serverInfo struct {
    weight   int
    disabled bool
}

type connList struct {
    count int
    head  *Conn
    tail  *Conn
}

type Pool struct {
    lock      sync.RWMutex
    hashRing  *Util.HashRing
    connLists map[string]*connList
    servers   map[string]*serverInfo

    prefix        string
    password      string
    db            int
    maxIdleConn   int
    maxIdleTime   time.Duration
    netTimeout    time.Duration
    probeInterval time.Duration
    mod           string
    modObj        IPool

    // 重新检查标志
    reCheck string
}

func (p *Pool) Construct() {
    p.hashRing = Util.NewHashRing()
    p.connLists = make(map[string]*connList)
    p.servers = make(map[string]*serverInfo)

    p.prefix = defaultPrefix
    p.password = defaultPassword
    p.db = defaultDb
    p.maxIdleConn = defaultIdleConn
    p.maxIdleTime = defaultIdleTime
    p.netTimeout = defaultTimeout
    p.probeInterval = defaultProbe
    p.mod = ModCluster
}

func (p *Pool) Init() {
    if len(p.servers) == 0 {
        p.servers[defaultServer] = &serverInfo{weight: 1, disabled: false}
    }

    if p.mod == ModCluster {
        p.modObj = newClusterPool(p).(IPool)
    } else {
        p.modObj = newMasterSlavePool(p).(IPool)
    }

    // first init cluster/master-slaves
    p.modObj.startCheck()

    if p.probeInterval != 0 {
        if p.probeInterval > maxProbeInterval {
            p.probeInterval = maxProbeInterval
        } else if p.probeInterval < minProbeInterval {
            p.probeInterval = minProbeInterval
        }

        go p.probeLoop()
    }
}

func (p *Pool) SetPrefix(prefix string) {
    p.prefix = prefix
}

func (p *Pool) SetPassword(password string) {
    p.password = password
}

func (p *Pool) SetDb(db int) {
    p.db = db
}

func (p *Pool) SetServers(v []interface{}) {
    for _, vv := range v {
        addr := vv.(string)

        if pos := strings.Index(addr, "://"); pos != -1 {
            addr = addr[pos+3:]
        }

        info := p.servers[addr]
        if info == nil {
            info = &serverInfo{}
            p.servers[addr] = info
        }

        info.weight += 1
    }
}

func (p *Pool) GetServers() (servers []string) {
    for server := range p.servers {
        servers = append(servers, server)
    }
    return
}

func (p *Pool) SetMaxIdleConn(v int) {
    p.maxIdleConn = v
}

func (p *Pool) SetMaxIdleTime(v string) {
    if maxIdleTime, e := time.ParseDuration(v); e != nil {
        panic(fmt.Sprintf(errSetProp, "maxIdleTime", e))
    } else {
        p.maxIdleTime = maxIdleTime
    }
}

func (p *Pool) SetNetTimeout(v string) {
    if netTimeout, e := time.ParseDuration(v); e != nil {
        panic(fmt.Sprintf(errSetProp, "netTimeout", e))
    } else {
        p.netTimeout = netTimeout
    }
}

func (p *Pool) SetProbeInterval(v string) {
    if probeInterval, e := time.ParseDuration(v); e != nil {
        panic(fmt.Sprintf(errSetProp, "probeInterval", e))
    } else {
        p.probeInterval = probeInterval
    }
}

func (p *Pool) SetMod(v string) {
    if Util.SliceSearchString(allMod, v) == -1 {
        panic("Undefined mod:" + v)
    }

    p.mod = v
}

func (p *Pool) BuildKey(key string) string {
    return p.prefix + key
}

func (p *Pool) AddrNewKeys(cmd string, v interface{}) (map[string][]string, map[string]string) {
    addrKeys, newKeys := make(map[string][]string), make(map[string]string)
    prevAddr := ""
    switch vv := v.(type) {
    case []string:
        for _, key := range vv {
            newKey := p.BuildKey(key)
            addr := p.GetAddrByKey(cmd, newKey, prevAddr)
            prevAddr = addr
            newKeys[newKey] = key
            addrKeys[addr] = append(addrKeys[addr], newKey)
        }
    case map[string]interface{}:
        for key := range vv {
            newKey := p.BuildKey(key)
            addr := p.GetAddrByKey(cmd, newKey, prevAddr)
            prevAddr = addr
            newKeys[newKey] = key
            addrKeys[addr] = append(addrKeys[addr], newKey)
        }
    default:
        panic(errBase + "addr new keys invalid")
    }
    return addrKeys, newKeys
}

func (p *Pool) RunAddrFunc(addr string, keys []string, wg *sync.WaitGroup, f func(*Conn, []string)) {
    defer func() {
        if err := recover(); err != nil {
            pgo.GLogger().Error("go coroutine RunAddrFunc,err:" + Util.ToString(err))
        }
        wg.Done() // notify done
    }()

    conn := p.GetConnByAddr(addr)
    defer conn.Close(false)

    f(conn, keys)
}

func (p *Pool) GetConnByKey(cmd, key string) *Conn {
    if addr := p.GetAddrByKey(cmd, key); len(addr) == 0 {
        panic(errNoServer)
    } else {
        return p.GetConnByAddr(addr)
    }
}

func (p *Pool) GetConnByAddr(addr string) *Conn {
    conn := p.getFreeConn(addr)
    if conn == nil || !p.checkConn(conn) {
        conn = p.dial(addr)
    }

    conn.ExtendDeadLine()
    return conn
}

// get redis address/node
// prevDft 一般用于master-slave mset mget mdel
func (p *Pool) GetAddrByKey(cmd, key string, prevDft ...string) string {
    cmd = strings.ToUpper(cmd)
    prev := ""
    if len(prevDft) > 0 {
        prev = prevDft[0]
    }
    return p.modObj.getAddrByKey(cmd, key, prev)
}

func (p *Pool) getFreeConn(addr string) *Conn {
    p.lock.Lock()
    defer p.lock.Unlock()
    list := p.connLists[addr]
    if list == nil || list.count == 0 {
        return nil
    }

    conn := list.head
    if list.count--; list.count == 0 {
        list.head, list.tail = nil, nil
    } else {
        list.head, conn.next.prev = conn.next, nil
    }

    conn.next = nil
    return conn
}

func (p *Pool) putFreeConn(conn *Conn) bool {
    p.lock.Lock()
    defer p.lock.Unlock()

    list := p.connLists[conn.addr]
    if list == nil {
        list = new(connList)
        p.connLists[conn.addr] = list
    }

    if list.count >= p.maxIdleConn {
        return false
    }

    if list.count == 0 {
        list.head, list.tail = conn, conn
        conn.prev, conn.next = nil, nil
    } else {
        conn.prev, conn.next = list.tail, nil
        conn.prev.next, list.tail = conn, conn
    }

    list.count++
    return true
}

func (p *Pool) checkConn(conn *Conn) bool {
    defer func() {
        // if panic, return value is default(false)
        if v := recover(); v != nil {
            conn.Close(true)
        }
    }()

    if !conn.CheckActive() {
        conn.Close(true)
        return false
    }
    return true
}

func (p *Pool) dial(addr string) *Conn {
    if nc, e := net.DialTimeout(p.parseNetwork(addr), addr, p.netTimeout); e != nil {
        panic(errBase + e.Error())
    } else {
        conn := newConn(addr, nc, p)
        defer func() {
            if v := recover(); v != nil {
                conn.Close(true)
                panic(v)
            }
        }()

        if len(p.password) > 0 {
            conn.Do("AUTH", p.password)
        }

        if p.db > 0 {
            conn.Do("SELECT", p.db)
        }

        return conn
    }
}

func (p *Pool) parseNetwork(addr string) string {
    if pos := strings.IndexByte(addr, '/'); pos != -1 {
        return "unix"
    } else {
        return "tcp"
    }
}

func (p *Pool) probeServer(addr string) {
    nc, e := net.DialTimeout(p.parseNetwork(addr), addr, p.netTimeout)
    if e != nil && !p.servers[addr].disabled {
        p.setServerDisabled(addr, true)
        p.modObj.check(addr, NodeActionDel)
    } else if e == nil && p.servers[addr].disabled {
        p.setServerDisabled(addr, false)
        p.modObj.check(addr, NodeActionAdd)

    }

    if e == nil {
        err := nc.Close()
        if err != nil {

        }
    }

    // 强制重新检查master
    if p.reCheck != "" {
        p.modObj.check(p.reCheck, NodeActionDel)
    }
}

func (p *Pool) setServerDisabled(addr string, disabled bool) {
    p.lock.Lock()
    defer p.lock.Unlock()
    p.servers[addr].disabled = disabled
}

func (p *Pool) probeLoop() {
    for {
        <-time.After(p.probeInterval)
        for addr := range p.servers {
            p.probeServer(addr)
        }
    }
}
