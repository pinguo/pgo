package Db

import (
	"gopkg.in/mgo.v2"
	"github.com/gomodule/redigo/redis"
)

type IMongoSession interface {
	Session() *mgo.Session
}

type IRedisConnect interface {
	RedisPool(key string) (*redis.Pool, error)
	GetKeyPrefix() string
}