package Memory

import (
    "time"

    "github.com/pinguo/pgo"
)

const (
    AdapterClass = "@pgo/Client/Memory/Adapter"

    defaultComponentId = "memory"
    defaultGcMaxItems  = 1000
    defaultGcInterval  = 60 * time.Second
    defaultExpire      = 60 * time.Second

    minGcInterval = 10 * time.Second
    maxGcInterval = 600 * time.Second

    errSetProp = "memory: failed to set %s, %s"
)

func init() {
    container := pgo.App.GetContainer()

    container.Bind(&Adapter{})
    container.Bind(&Client{})
}
