package Db

import (
    "sync"
    "time"

    "github.com/pinguo/pgo"
)

const (
    defaultComponentId = "db"
    defaultTimeout     = 10 * time.Second
)

var stmtPool sync.Pool

func init() {
    container := pgo.App.GetContainer()

    container.Bind(&Adapter{})
    container.Bind(&Client{})

    stmtPool.New = func() interface{} {
        return &Stmt{}
    }
}
