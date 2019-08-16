package RabbitMq

import (
    "fmt"
    "net"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/pinguo/pgo"
    "github.com/pinguo/pgo/Util"
)

type serverInfo struct {
    weight int64
}

type Pool struct {
    ServiceName string
    servers     map[string]*serverInfo
    tlsRootCAs  string
    tlsCert     string
    tlsCertKey  string
    user        string
    pass        string

    exchangeName string
    exchangeType string

    maxChannelNum      int
    maxIdleChannel     int
    maxIdleChannelTime time.Duration
    maxWaitTime        time.Duration

    probeInterval time.Duration

    connList map[string]*ConnBox

    lock sync.RWMutex
}

func (c *Pool) Construct() {
    c.connList = make(map[string]*ConnBox)
    c.servers = make(map[string]*serverInfo)
    c.maxChannelNum = dftMaxChannelNum
    c.maxIdleChannel = dftMaxIdleChannel
    c.maxIdleChannelTime = dftMaxIdleChannelTime
    c.exchangeType = dftExchangeType
    c.exchangeName = dftExchangeName
    c.maxWaitTime = dftMaxWaitTime
    c.probeInterval = dftProbeInterval

}

func (c *Pool) Init() {
    if c.exchangeName == "" {
        panic("exchangeName cannot be empty")
    }

    if c.ServiceName == "" {
        panic("ServiceName cannot be empty")
    }

    if c.maxIdleChannel > c.maxChannelNum {
        panic("maxIdleChannel cannot be larger than maxChannelNum")
    }

    if c.probeInterval > 0 {
        go c.probeLoop()
    }

}

func (c *Pool) SetServers(v []interface{}) {
    for _, vv := range v {
        addr := vv.(string)

        if pos := strings.Index(addr, "://"); pos != -1 {
            addr = addr[pos+3:]
        }

        info := c.servers[addr]
        if info == nil {
            info = &serverInfo{}
            c.servers[addr] = info
        }

        info.weight += 1
    }
}

func (c *Pool) GetServers() (servers []string) {
    for server := range c.servers {
        servers = append(servers, server)
    }
    return servers
}

func (c *Pool) SetUser(v string) {
    c.user = v
}

func (c *Pool) SetPass(v string) {
    c.pass = v
}

func (c *Pool) SetTlsRootCAs(v string) {
    c.tlsRootCAs = v
}

func (c *Pool) SetTlsCert(v string) {
    c.tlsCert = v
}

func (c *Pool) SetTlsCertKey(v string) {
    c.tlsCertKey = v
}

func (c *Pool) SetExchangeName(v string) {
    c.exchangeName = v
}

func (c *Pool) SetServiceName(v string) {
    c.ServiceName = v
}

func (c *Pool) SetExchangeType(v string) {
    c.exchangeType = v
}

func (c *Pool) SetMaxChannelNum(v int) {
    c.maxChannelNum = v
}

func (c *Pool) SetMaxIdleChannel(v int) {
    c.maxIdleChannel = v
}

func (c *Pool) setMaxIdleChannelTime(v string) {
    if netTimeout, e := time.ParseDuration(v); e != nil {
        panic(fmt.Sprintf(errSetProp, "maxIdleChannelTime", e))
    } else {
        c.maxIdleChannelTime = netTimeout
    }
}

func (c *Pool) setMaxWaitTime(v string) {
    if netTimeout, e := time.ParseDuration(v); e != nil || netTimeout <= 0 {
        panic(fmt.Sprintf(errSetProp, "maxWaitTime", e))
    } else {
        c.maxWaitTime = netTimeout
    }
}

func (c *Pool) SetProbeInterval(v string) {
    if probeInterval, e := time.ParseDuration(v); e != nil {
        panic(fmt.Sprintf(errSetProp, "probeInterval", e))
    } else {
        c.probeInterval = probeInterval
    }
}

func (c *Pool) getExchangeName() string {
    return "pgo." + c.exchangeName
}

func (c *Pool) getRouteKey(opCode string) string {
    return "pgo." + c.exchangeName + "." + opCode
}

// 获取channel链接
func (c *Pool) getFreeChannel() *ChannelBox {
    connBox := c.getConnBox()
    connBox.useChannelCount++

    var channelBox *ChannelBox

    select {
    case channelBox = <-connBox.channelList:
    default:
    }

    if channelBox == nil {
        return c.getChannelBox(connBox)
    }

    if time.Since(channelBox.lastActive) >= c.maxIdleChannelTime || channelBox.connStartTime != connBox.startTime {
        channelBox.Close(true)
        return c.getChannelBox(connBox)
    }
    return channelBox
}

// 获取ChannelBox
func (c *Pool) getChannelBox(connBox *ConnBox) *ChannelBox {
    if connBox.useConnCount >= c.maxChannelNum {
        // 等待回收
        var channelBox *ChannelBox
        timeAfter := time.After(c.maxWaitTime)
        select {
        case channelBox = <-connBox.channelList:
        case <-timeAfter:
        }

        if channelBox == nil {
            panic("RabbitMq gets the channel timeout")
        }

        return channelBox
    } else {
        return newChannelBox(connBox, c)
    }
}

// 释放或者返回channel链接池
func (c *Pool) putFreeChannel(channelBox *ChannelBox) bool {
    connBox := c.getConnBox(channelBox.connBoxId)

    if len(connBox.channelList) >= c.maxIdleChannel {
        connBox.useChannelCount--
        return false
    }

    select {
    case connBox.channelList <- channelBox:
    default:
    }

    return true

}

// 获取tcp链接
func (c *Pool) getConnBox(idDft ...string) *ConnBox {
    if len(c.connList) == 0 {
        c.initConn()
    }

    c.lock.RLock()
    defer c.lock.RUnlock()
    if len(idDft) > 0 {
        return c.connList[idDft[0]]
    }

    k := ""
    num := 0
    for i, connBox := range c.connList {
        if connBox.isClosed() {
            continue
        }
        cLen := len(connBox.channelList)
        if num == 0 || cLen > num {
            k = i
            num = cLen
        }
    }
    if k == "" {
        panic("Rabbit not found conn")
    }
    return c.connList[k]
}

// 设置tcp链接
func (c *Pool) initConn() bool {

    c.lock.Lock()
    defer c.lock.Unlock()
    for addr, info := range c.servers {
        var i int64
        for i = 1; i <= info.weight; i++ {
            id := c.getConnId(addr, i)
            c.connList[id] = newConnBox(id, addr, c.getDsn(addr), c.maxChannelNum, c.tlsCert, c.tlsCertKey, c.tlsRootCAs)
        }

    }

    return true
}

func (c *Pool) getConnId(addr string, i int64) string {
    return addr + "_" + strconv.FormatInt(i, 10)
}

func (c *Pool) getDsn(addr string) string {
    return fmt.Sprintf("%s://%s:%s@%s", dftProtocol, c.user, c.pass, addr)
}

func (c *Pool) probeServer(addr string, weight int64) {
    nc, e := net.DialTimeout("tcp", addr, defaultTimeout)
    if e == nil {
        defer nc.Close()
    }

    var i int64
    for i = 1; i <= weight; i++ {
        id := c.getConnId(addr, i)
        connBox := c.getConnBox(id)
        func() {
            defer func() {
                if err := recover(); err != nil {
                    pgo.GLogger().Error("Rabbit probeServer err:" + Util.ToString(err))
                }
            }()

            if e != nil && !connBox.isClosed() {
                connBox.setDisable()
            } else if e == nil && connBox.isClosed() {
                connBox.setEnable()
                connBox.initConn()
            }
        }()

    }
}

func (c *Pool) probeLoop() {
    for {
        <-time.After(c.probeInterval)
        for addr, info := range c.servers {
            c.probeServer(addr, info.weight)
        }
    }
}
