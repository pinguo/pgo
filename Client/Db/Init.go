package Db

import (
    "time"

    "github.com/pinguo/pgo"
)

const (
    defaultComponentId = "db"
    defaultTimeout     = 10 * time.Second
)

func init() {
    container := pgo.App.GetContainer()

    container.Bind(&Adapter{})
    container.Bind(&Client{})
}
