package Db

import (
    "github.com/gomodule/redigo/redis"
    "gopkg.in/mgo.v2"
)

type IMongoSession interface {
    Session() *mgo.Session
}

type IRedisConnect interface {
    RedisPool(key string) (*redis.Pool, error)
    GetKeyPrefix() string
}
