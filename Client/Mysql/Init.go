package Mysql

import "github.com/pinguo/pgo"

func init() {
    container := pgo.App.GetContainer()
    container.Bind(&Adapter{})
    container.Bind(&Client{})
}
