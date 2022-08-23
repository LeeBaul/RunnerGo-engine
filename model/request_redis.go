package model

import (
	"github.com/go-redis/redis"
	"time"
)

var (
	RDB          *redis.Client
	timeDuration = 3 * time.Second
)

func InitRedisClient(addr, password string, db, size int64) (err error) {
	RDB = redis.NewClient(
		&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       int(db),
			PoolSize: int(size),
		})
	_, err = RDB.Ping().Result()
	return err
}

func QueryPlanStatus(key string, ch chan bool) {
	defer close(ch)
	ticker := time.NewTicker(timeDuration)
	for {
		select {
		case <-ticker.C:
			value, _ := RDB.Get(key).Result()
			if value == "false" {
				ch <- false
				ticker.Stop()
				return
			}
		}
	}
}

// QueryTimingTaskStatus 查询定时任务状态
func QueryTimingTaskStatus(key string) bool {
	ticker := time.NewTicker(timeDuration)
	for {
		select {
		case <-ticker.C:
			value, _ := RDB.Get(key).Result()
			if value == "false" {
				return false
			}
		}
	}
}
