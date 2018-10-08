package Mongo

import (
    "time"

    "github.com/pinguo/pgo"
)

const (
    defaultDsn     = "mongodb://127.0.0.1:27017/"
    defaultOptions = "connect=replicaSet&maxPoolSize=100&minPoolSize=1&maxIdleTimeMS=300000" +
        "&ssl=false&w=1&j=false&wtimeoutMS=10000&readPreference=secondaryPreferred"

    defaultComponentId    = "mongo"
    defaultConnectTimeout = 1 * time.Second
    defaultReadTimeout    = 10 * time.Second
    defaultWriteTimeout   = 10 * time.Second

    errSetProp    = "mongo: failed to set %s, %s"
    errInvalidDsn = "mongo: invalid dsn %s, %s"
    errInvalidOpt = "mongo: invalid option "
    errDialFailed = "mongo: failed to dial %s, %s"
)

func init() {
    container := pgo.App.GetContainer()

    container.Bind(&Adapter{})
    container.Bind(&Client{})
}
