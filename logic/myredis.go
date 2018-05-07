package logic

import (
	"time"

	"im_msg_server/common"

	"github.com/gomodule/redigo/redis"
)

const (
	REDIS_T_HASH    = 1
	REDIS_T_SET     = 2
	REDIS_T_KEYS    = 3
	REDIS_T_STRING  = 4
	REDIS_T_LIST    = 5
	REDIS_T_SORTSET = 6
)

const (
	MGO_IDLE_COUNT   = 1   //连接池空闲个数
	MGO_ACTIVE_COUNT = 10  //连接池活动个数
	MGO_IDLE_TIMEOUT = 180 //空闲超时时间
)

var RedisClients map[int]*redis.Pool

func init() {
	RedisClients = make(map[int]*redis.Pool)
	str := common.Conf.GetKeyList("queue")
	for index, value := range str {
		if str1, err := common.Conf.GetValue("queue", value); nil == err {
			RedisClients[index] = createPool(MGO_IDLE_COUNT, MGO_ACTIVE_COUNT, MGO_IDLE_TIMEOUT, str1)
			index++
		}
	}
}

func createPool(maxIdle, maxActive, idleTimeout int, address string) (obj *redis.Pool) {
	obj = new(redis.Pool)
	obj.MaxIdle = maxIdle
	obj.MaxActive = maxActive
	obj.IdleTimeout = (time.Duration)(idleTimeout) * time.Second
	obj.Dial = func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", address)
		if err != nil {
			return nil, err
		}
		return c, err
	}
	return
}

func GetQueueConn() (conn redis.Conn) {
	if len(RedisClients) <= 0 {
		return nil
	}

	return RedisClients[0].Get()
}
