package Redis

import (
    "time"

    "github.com/pinguo/pgo"
)

const (
    AdapterClass = "@pgo/Client/Redis/Adapter"

    defaultComponentId = "redis"
    defaultServer      = "127.0.0.1:6379"
    defaultPrefix      = "pgo_"
    defaultPassword    = ""
    defaultDb          = 0
    defaultProbe       = 0
    defaultIdleConn    = 10
    defaultIdleTime    = 60 * time.Second
    defaultTimeout     = 1 * time.Second
    defaultExpire      = 24 * time.Hour

    ModCluster         = "cluster"
    ModMasterSlave     = "masterSlave"


    maxProbeInterval = 30 * time.Second
    minProbeInterval = 1 * time.Second

    errBase        = "redis: "
    errSetProp     = "redis: failed to set %s, %s"
    errNoServer    = "redis: no server available"
    errInvalidResp = "redis: invalid resp type, "
    errSendFailed  = "redis: send request failed, "
    errReadFailed  = "redis: read response failed, "
    errCorrupted   = "redis: corrupted response, "

    PgoMasterSlaveCheckPrefix = "pgo_master_slave_check_"

    NodeActionAdd = "add"
    NodeActionDel = "del"
)

var (
    lineEnding = []byte("\r\n")
    replyOK    = []byte("OK")
    replyPong  = []byte("PONG")
    allMod     = []string{ModCluster, ModMasterSlave}

    allRedisCmd = []string{
        // Strings
        "DECR", "DECRBY", "GETSET", "INCR", "INCRBY", "INCRBYFLOAT",
        "SETEX", "PSETEX", "SETNX",

        // Keys
        "EXISTS", "EXPIRE", "PEXPIRE", "EXPIREAT", "PEXPIREAT",
        "PERSIST", "RENAME", "RENAMENX", "TYPE", "TTL",  "PTTL",

        // Hashes
        "HDEL", "HEXISTS", "HGET", "HGETALL", "HINCRBY", "HINCRBYFLOAT", "HKEYS",
        "HLEN", "HMGET", "HMSET", "HSET","HSETNX", "HVALS",

        // List
        "BLPOP", "BRPOP", "LINDEX", "LGET", "LINSERT", "LLEN",
        "LPOP", "LPUSH", "LPUSHX", "LRANGE", "LREM", "LGETRANGE", "LSET",
        "LTRIM", "RPOP", "RPUSH", "RPUSHX",

        // Set
        "SADD", "SCARD", "SISMEMBER","SMEMBERS", "SPOP", "SRANDMEMBER", "SREM",

        // Sorted Set
        "ZADD", "ZCARD", "ZCOUNT", "ZINCRBY", "ZRANGE", "ZRANGEBYSCORE", "ZREVRANGEBYSCORE",
        "ZRANK", "ZREVRANK", "ZREM", "ZREMRANGEBYRANK", "ZREMRANGEBYSCORE", "ZREVRANGE", "ZSCORE",
    }

    allRedisReadCmd = []string{
        // Strings
        "GET", "MGET", "BITCOUNT", "STRLEN", "GETBIT", "GETRANGE",
        // Keys
        "KEYS", "TYPE", "SCAN", "EXISTS", "PTTL", "TTL",
        // Hashes
        "HEXISTS", "HGETALL", "HKEYS", "HLEN", "HGET", "HMGET",
        // Set
        "SISMEMBER", "SMEMBERS", "SRANDMEMBER", "SSCAN", "SCARD", "SDIFF", "SINTER",
        // List
        "LINDEX", "LLEN", "LRANGE",
        // Sorted Set
        "ZCARD", "ZCOUNT", "ZRANGE", "ZRANGEBYSCORE", "ZRANK", "ZREVRANGE", "ZREVRANGEBYSCORE",
        "ZREVRANK", "ZSCAN", "ZSCORE",
    }
)

func init() {
    container := pgo.App.GetContainer()

    container.Bind(&Adapter{})
    container.Bind(&Client{})

}

func keys2Args(keys []string) []interface{} {
    args := make([]interface{}, len(keys))
    for i, k := range keys {
        args[i] = k
    }
    return args
}

type IPool interface {
    startCheck()
    check(addr, aType string)
    getAddrByKey(cmd, key, prevDft string) string
}
