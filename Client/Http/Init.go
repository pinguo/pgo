package Http

import (
    "net/url"
    "time"

    "github.com/pinguo/pgo"
)

const (
    defaultComponentId = "http"
    defaultUserAgent   = "PGO Framework"
    defaultTimeout     = 10 * time.Second
)

func init() {
    container := pgo.App.GetContainer()

    container.Bind(&Adapter{})
    container.Bind(&Client{})
}

func baseUrl(addr string) string {
    u, e := url.Parse(addr)
    if e != nil {
        panic("http parse url failed, " + e.Error())
    }

    return u.Scheme + "://" + u.Host + u.Path
}
