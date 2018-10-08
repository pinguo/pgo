package Db

import (
	"fmt"
	"github.com/pinguo/pgo/Util"
	"github.com/pinguo/pgo"
)

type RedisProxy struct {
	pgo.Object
	componentId string
}

func (r *RedisProxy) Construct(componentId string){
	r.componentId = componentId
}

// 实行具体命令
func (r *RedisProxy) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	key, err := Util.ToBytes(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid key %v", args[0])
	}
    redisPool, err := pgo.App.Get(r.componentId).(IRedisConnect).RedisPool(key)
	if err != nil {
		return nil, fmt.Errorf("get redis pool err msg %v", err)
	}
	logKey := "redis." + commandName
	r.GetContext().ProfileStart(logKey)
	defer r.GetContext().ProfileStop(logKey)
    args[0] = pgo.App.Get(r.componentId).(IRedisConnect).GetKeyPrefix() + key
	redisC := redisPool.Get()
	defer redisC.Close()
	return redisC.Do(commandName, args...)
}

