package Memcache

import (
    "time"

    "github.com/pinguo/pgo"
)

const (
    defaultServer      = "127.0.0.1:11211"
    defaultComponentId = "memcache"
    defaultPrefix      = "pgo_"
    defaultProbe       = 0
    defaultIdleConn    = 10
    defaultIdleTime    = 60 * time.Second
    defaultTimeout     = 1 * time.Second
    defaultExpire      = 24 * time.Hour

    maxProbeInterval = 30 * time.Second
    minProbeInterval = 1 * time.Second

    CmdCas      = "cas"
    CmdAdd      = "add"
    CmdSet      = "set"
    CmdReplace  = "replace"
    CmdAppend   = "append"
    CmdPrepend  = "prepend"
    CmdGet      = "get"
    CmdGets     = "gets"
    CmdDelete   = "delete"
    CmdIncr     = "incr"
    CmdDecr     = "decr"
    CmdTouch    = "touch"
    CmdStats    = "stats"
    CmdFlushAll = "flush_all"
    CmdVersion  = "version"

    errBase       = "memcache: "
    errSetProp    = "memcache: failed to set %s, %s"
    errNoServer   = "memcache: no server available"
    errInvalidCmd = "memcache: invalid cmd, "
    errSendFailed = "memcache: send request failed, "
    errReadFailed = "memcache: read response failed, "
    errEmptyKeys  = "memcache: empty keys"
    errCorrupted  = "memcache: corrupted response, "
)

var (
    lineEnding     = []byte("\r\n")
    replyOK        = []byte("OK\r\n")
    replyStored    = []byte("STORED\r\n")
    replyNotStored = []byte("NOT_STORED\r\n")
    replyExists    = []byte("EXISTS\r\n")
    replyNotFound  = []byte("NOT_FOUND\r\n")
    replyDeleted   = []byte("DELETED\r\n")
    replyTouched   = []byte("TOUCHED\r\n")
    replyEnd       = []byte("END\r\n")
)

func init() {
    container := pgo.App.GetContainer()

    container.Bind(&Adapter{})
    container.Bind(&Client{})
}
