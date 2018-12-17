package Db

import (
    "sync"
    "time"

    "github.com/pinguo/pgo"
)

const (
    AdapterClass = "@pgo/Client/Db/Adapter"

    defaultComponentId = "db"
    defaultTimeout     = 10 * time.Second
)

var (
    stmtPool sync.Pool
    rowPool  sync.Pool
)

func init() {
    container := pgo.App.GetContainer()

    container.Bind(&Adapter{})
    container.Bind(&Client{})

    stmtPool.New = func() interface{} { return &Stmt{} }
    rowPool.New = func() interface{} { return &Row{} }
}
