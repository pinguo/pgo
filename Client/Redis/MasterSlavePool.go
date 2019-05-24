package Redis

import (
    "bytes"
    "math/rand"
    "os"
    "strings"
    "time"

    "github.com/pinguo/pgo"
    "github.com/pinguo/pgo/Util"
)

type MasterSlavePool struct {
    pool *Pool

    master string
    slaves []string
}

func newMasterSlavePool(pool *Pool) interface{} {
    return &MasterSlavePool{
        pool: pool,
    }
}

// prevDft 上一个默认addr  一般在master-slave用于mset mget mdel
func (m *MasterSlavePool) getAddrByKey(cmd, key, prevDft string) string {
    if prevDft != "" {
        return prevDft
    }
    m.pool.lock.RLock()
    defer m.pool.lock.RUnlock()
    if Util.SliceSearchString(allRedisReadCmd, cmd) >= 0 && len(m.slaves) > 0 {
        rand.Seed(time.Now().UnixNano())
        index := rand.Intn(len(m.slaves))
        return m.slaves[index]
    } else {
        return m.master
    }
}

// timing check master salve
func (m *MasterSlavePool) check(addr, aType string) {
    oldMaster := m.master
    oldSlaves := m.slaves

    if addr == m.master {
        // is master
        newMaster := m.startCheckMaster()
        if newMaster == "" {
            // 未找到master 强制重找
            m.pool.reCheck = addr
        } else {
            m.pool.reCheck = ""
        }
        m.delSlave(m.master)
        m.delSlave(addr)

    } else {
        // is slave
        serverInfo := m.pool.servers[addr]

        for i := 0; i < serverInfo.weight; i++ {
            if aType == NodeActionAdd {
                m.slaves = append(m.slaves, addr)
                m.delSlave(m.master)
            } else if aType == NodeActionDel {
                m.delSlave(addr)
            }
        }

    }

    if len(m.slaves) == 0 {
        // check master
        m.startCheckMaster()

        // check  slave
        m.startCheckSlaves()
    }

    if m.master == "" {
        pgo.GLogger().Error("Timing check,No master redis server in master-slave config!")
        return
    }

    if len(m.slaves) == 0 {
        pgo.GLogger().Error("Timing check,No slaves redis server in master-slave config!")
        return
    }

    if oldMaster != m.master {
        pgo.GLogger().Warn("Redis Proxy master node change, " + oldMaster + " to " + m.master)
    }

    diffSlaves := make([]string, 0)
    for _, v := range oldSlaves {
        if Util.SliceSearchString(m.slaves, v) == -1 {
            diffSlaves = append(diffSlaves, v)
        }
    }

    if len(diffSlaves) > 0 {
        pgo.GLogger().Warn("Redis Proxy slave nodes change , (" + strings.Join(oldSlaves, ",") + ") to (" + strings.Join(m.slaves, ",") + ")")
    }

}

func (m *MasterSlavePool) delSlave(addr string) {

    if index := Util.SliceSearchString(m.slaves, addr); index >= 0 {
        m.pool.lock.Lock()
        defer m.pool.lock.Unlock()
        m.slaves = append(m.slaves[0:index], m.slaves[index+1:]...)
    }
}

// first check master and check slave
func (m *MasterSlavePool) startCheck() {
    // check master
    m.startCheckMaster()
    // check  slave
    m.startCheckSlaves()

    if m.master == "" {
        panic("No master redis server in master-slave config!")
    }

    if len(m.slaves) == 0 {
        panic("No slave redis server in master-slave config!")
    }

    return
}

func (m *MasterSlavePool) startCheckMaster() string {
    master := ""
    for addr, serverInfo := range m.pool.servers {
        if serverInfo.disabled == true {
            continue
        }

        master = m.checkMaster(addr)
        if master != "" {
            break
        }
    }
    return master
}

func (m *MasterSlavePool) startCheckSlaves() {
    // check slaves
    if len(m.pool.servers) == 1 {
        m.slaves = append(m.slaves, m.master)
    } else {
        for addr, serverInfo := range m.pool.servers {
            if serverInfo.disabled == true {
                continue
            }

            if addr == m.master {
                continue
            }

            m.checkSlave(addr)
        }
    }

    // only master active
    if len(m.slaves) == 0 {
        m.checkSlave(m.master)
    }
}

// check master
func (m *MasterSlavePool) checkMaster(addr string) string {
    defer func() {
        recover()
    }()

    conn := m.pool.GetConnByAddr(addr)
    defer conn.Close(false)
    res := conn.Do("SET", m.getCheckKey(), 1, "EX", 5*time.Second)
    payload, ok := res.([]byte)
    if ok && bytes.Equal(payload, replyOK) {
        func() {
            m.pool.lock.Lock()
            defer m.pool.lock.Unlock()
            m.master = addr
        }()

        return addr
    }

    return ""
}

// check slave
func (m *MasterSlavePool) checkSlave(addr string) string {
    defer func() {
        recover()
    }()

    conn := m.pool.GetConnByAddr(addr)
    defer conn.Close(false)

    conn.Do("GET", m.getCheckKey())
    serverInfo := m.pool.servers[addr]
    for i := 0; i < serverInfo.weight; i++ {
        m.slaves = append(m.slaves, addr)
    }

    return addr
    //ret := pgo.NewValue(retI)
    //if ret.Int() == 1 {
    //	serverInfo := m.pool.servers[addr]
    //	for i := 0; i < serverInfo.weight; i++ {
    //		m.slaves = append(m.slaves, addr)
    //	}
    //
    //	return addr
    //}
    //
    //return ""
}

func (m *MasterSlavePool) getCheckKey() string {
    hostName, _ := os.Hostname()
    return PgoMasterSlaveCheckPrefix + hostName
}
