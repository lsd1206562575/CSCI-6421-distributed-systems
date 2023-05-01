package redis

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

var (
	pool      *redis.Pool
	redisHost = "127.0.0.1:6379"
)

func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10, // 最大空闲连接数
		MaxActive:   0,  // 最大激活连接数（0 表示没有限制）
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", redisHost)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
}

func init() {
	pool = newRedisPool()
}

func RedisPool() *redis.Pool {
	return pool
}
