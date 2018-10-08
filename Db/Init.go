package Db

import "pgo"

func init() {
    container := pgo.App.GetContainer()

    container.Bind(&MongoSession{})
    container.Bind(&RedisConnect{})
    container.Bind(&RedisProxy{})
    container.Bind(&MongoProxy{})
}
