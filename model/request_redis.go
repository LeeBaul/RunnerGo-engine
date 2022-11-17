package model

import (
	"github.com/go-redis/redis"
	"time"
)

var (
	RDB          *redis.Client
	timeDuration = 3 * time.Second
)

type RedisClient struct {
	Client *redis.Client
}

func InitRedisClient(addr, password string, db int64) (err error) {
	RDB = redis.NewClient(
		&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       int(db),
		})
	_, err = RDB.Ping().Result()
	return err
}

func InsertStatus(key, value string, expiration time.Duration) (err error) {
	if expiration < 20*time.Second {
		expiration = 20 * time.Second
	}
	err = RDB.Set(key, value, expiration).Err()
	if err != nil {
		return
	}
	return
}

func QueryPlanStatus(key string) (err error, value string) {
	value, err = RDB.Get(key).Result()
	return
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
		time.Sleep(timeDuration)
	}
}

func QuerySceneStatus(key string) (err error, value string) {
	value, err = RDB.Get(key).Result()
	return
}

func QueryReportData(key string) (value string) {
	values := RDB.LRange(key, 0, -1).Val()
	if len(values) <= 0 {
		return
	}
	value = values[0]
	return
}

func InsertHeartbeat(key string, field string, value interface{}) error {
	err := RDB.HSet(key, field, value).Err()
	return err
}

func DeleteKey(key string) {
	_ = RDB.Del(key).Err()
	return
}

func InsertMachineResources(key string, value interface{}) error {
	err := RDB.LPush(key, value).Err()
	return err
}
