package Mysql

import "pgo"

func init() {
    container := pgo.App.GetContainer()
    container.Bind(&Adapter{})
    container.Bind(&Client{})
}
