package Redis

func newClusterPool(pool *Pool) interface{} {
    return &ClusterPool{
        pool: pool,
    }
}

type ClusterPool struct {
    pool *Pool
}

func (c *ClusterPool) getAddrByKey(cmd, key, prevDft string) string {
    c.pool.lock.RLock()
    defer c.pool.lock.RUnlock()
    return c.pool.hashRing.GetNode(key)
}

// first init Cluster
func (c *ClusterPool) startCheck() {
    for addr, item := range c.pool.servers {
        c.pool.hashRing.AddNode(addr, item.weight)
    }
}

// timing check cluster
func (c *ClusterPool) check(addr, aType string) {
    c.pool.lock.Lock()
    defer c.pool.lock.Unlock()
    if aType == NodeActionAdd {
        c.pool.hashRing.AddNode(addr, c.pool.servers[addr].weight)
    } else if aType == NodeActionDel {
        c.pool.hashRing.DelNode(addr)
    }
}
