package Redis

import (
    "pgo"
    "time"
)

const (
    defaultServer      = "127.0.0.1:6379"
    defaultComponentId = "redis"
    defaultPrefix      = "pgo_"
    defaultPassword    = ""
    defaultDb          = 0
    defaultProbe       = 0
    defaultIdleConn    = 10
    defaultIdleTime    = 60 * time.Second
    defaultTimeout     = 1 * time.Second
    defaultExpire      = 24 * time.Hour

    maxProbeInterval = 30 * time.Second
    minProbeInterval = 1 * time.Second

    errBase        = "redis: "
    errSetProp     = "redis: failed to set %s, %s"
    errNoServer    = "redis: no server available"
    errInvalidResp = "redis: invalid resp type, "
    errSendFailed  = "redis: send request failed, "
    errReadFailed  = "redis: read response failed, "
    errCorrupted   = "redis: corrupted response, "
)

var (
    lineEnding = []byte("\r\n")
    replyOK    = []byte("OK")
    replyPong  = []byte("PONG")
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
