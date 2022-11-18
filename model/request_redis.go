package model

import (
	"github.com/go-redis/redis"
	"time"
)

var (
	RDB1         *redis.Client
	RDB2         *redis.Client
	timeDuration = 3 * time.Second
)

type RedisClient struct {
	Client *redis.Client
}

func InitRedisClient(addr1, password1 string, db1 int64, addr2, password2 string, db2 int64) (err error) {
	RDB1 = redis.NewClient(
		&redis.Options{
			Addr:     addr1,
			Password: password1,
			DB:       int(db1),
		})
	_, err = RDB1.Ping().Result()
	if err != nil {
		return
	}
	RDB2 = redis.NewClient(
		&redis.Options{
			Addr:     addr2,
			Password: password2,
			DB:       int(db2),
		})
	_, err = RDB2.Ping().Result()
	return err
}

func QueryPlanStatus(key string) (err error, value string) {
	value, err = RDB1.Get(key).Result()
	return
}

func QuerySceneStatus(key string) (err error, value string) {
	value, err = RDB1.Get(key).Result()
	return
}

func QueryReportData(key string) (value string) {
	values := RDB2.LRange(key, 0, -1).Val()
	if len(values) <= 0 {
		return
	}
	value = values[0]
	return
}

func InsertHeartbeat(key string, field string, value interface{}) error {
	err := RDB1.HSet(key, field, value).Err()
	return err
}

func DeleteKey(key string) (err error) {
	err = RDB1.Del(key).Err()
	return
}

func InsertMachineResources(key string, value interface{}) error {
	err := RDB1.LPush(key, value).Err()
	return err
}
