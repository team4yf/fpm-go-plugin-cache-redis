package rds

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestRedisIsSet(t *testing.T) {
	redisOptions := &redis.Options{
		Addr:     "localhost:6379",
		Password: "admin123",
		DB:       1,
		PoolSize: 10,
	}
	cli := redis.NewClient(redisOptions)

	redisCache := NewRedisCache("test", cli)

	var success bool
	var err error
	err = redisCache.SetInt("a", 1, 100*time.Second)
	assert.Nil(t, err, "setInt should not occur error")

	success, err = redisCache.IsSet("a")
	assert.Nil(t, err, "IsSet should not occur error")
	assert.Equal(t, true, success, "IsSet should return true")

	success, err = redisCache.Remove("a")
	assert.Nil(t, err, "Remove should not occur error")
	assert.Equal(t, true, success, "Remove should return true")

	success, err = redisCache.IsSet("a")
	assert.Nil(t, err, "IsSet should not occur error")
	assert.Equal(t, false, success, "IsSet should return false")

}
