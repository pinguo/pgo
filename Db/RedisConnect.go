package Db

import (
	"github.com/gomodule/redigo/redis"
	"github.com/pinguo/pgo/Util"
	"time"
)

type RedisConnect struct {
	hashring     *HashRing
	KeyPrefix    string
	servers      map[string]map[string]string
	Maxidle      int
	Maxactive    int
	Db           int
}

func (r *RedisConnect) Construct() {
	r.Maxidle = 100
	r.Maxactive = 100
	r.Db = 0
	r.KeyPrefix = ""
}

// 初始化
func (r *RedisConnect) Init() {
	r.hashring = NewHashRing(0)
	r.hashring.AddNodes(r.nodeWeight())
}

// 设置servers
func (r *RedisConnect) SetServers(servers map[string]interface{}) {
	r.servers = make(map[string]map[string]string)
	for k, v := range servers {
		v1, ok := v.(map[string]interface{})
		if ok == false {
			panic("redis config componentId:" + r.KeyPrefix + " servers err")
		}
		if _, ok := r.servers[k]; ok == false {
			r.servers[k] = make(map[string]string)
		}
		for kk, vv := range v1 {
			vv1, ok := vv.(string)
			if ok == false {
				panic("redis config componentId:" + r.KeyPrefix + "." + k + "." + kk + " is not string")
			}
			r.servers[k][kk] = vv1
		}
	}
}

// 获取key前缀
func (r *RedisConnect) GetKeyPrefix() string {
	return r.KeyPrefix
}

// nodeWeight 获取节点和权重
func (r *RedisConnect) nodeWeight() map[string]int {
	nodeWeight := make(map[string]int)
	for k, v := range r.servers {
		nodeWeight[k] = Util.ToInt(v["weight"])
	}
	return nodeWeight
}

// 返回或者生成链接池
func (r *RedisConnect) RedisPool(key string) (*redis.Pool, error) {
	node := r.hashring.GetNode(key)
	ip, ok := r.servers[node]["ip"]
	if ok == false {
		panic("redis config not found " + node + " key ip")
	}
	port, ok := r.servers[node]["port"]
	if ok == false {
		panic("redis config not found " + node + " key port")
	}
	redisHost := ip + ":" + port
	// idleTimeout :=r.app.GetConfig().DefaultInt64(configBaseKey + ".idleTimeout", 180)
	maxidle := 1
	if r.Maxidle > 0 {
		maxidle = r.Maxidle
	}
	maxactive := 1
	if r.Maxactive > 0 {
		maxactive = r.Maxactive
	}
	db := 1
	if r.Db > 0 {
		db = r.Db
	}
	RedisClient := &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle:     maxidle,
		MaxActive:   maxactive,
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				return nil, err
			}
			// 选择db
			c.Do("SELECT", db)
			return c, nil
		},
	}
	return RedisClient, nil
}
