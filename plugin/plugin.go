//Package plugin the plugin body
package plugin

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/team4yf/fpm-go-plugin-cache-redis/rds"
	"github.com/team4yf/yf-fpm-server-go/fpm"
)

var defaultCtx = context.Background()

type Options struct {
	Prefix string
	Addr   string
	Passwd string
	DB     int
	Pool   int
}

func init() {
	fpm.Register(func(app *fpm.Fpm) {

		options := &Options{
			Prefix: "",
			Addr:   "localhost:6379",
			Passwd: "admin123",
			DB:     1,
			Pool:   10,
		}

		if app.HasConfig("redis") {
			if err := app.FetchConfig("redis", options); err != nil {
				app.Logger.Errorf("Fetch Redis Config Error: %v", err)
				panic(err)
			}
		}

		app.Logger.Debugf("Redis Config: %v", options)
		redisOptions := &redis.Options{
			Addr:     options.Addr,
			Password: options.Passwd,
			DB:       options.DB,
			PoolSize: options.Pool,
		}
		cli := redis.NewClient(redisOptions)
		if _, err := cli.Ping(defaultCtx).Result(); err != nil {
			app.Logger.Errorf("Redis Cant Connect! Cause: %v", err)
			panic(err)
		}

		app.SetCacher(rds.NewRedisCache(options.Prefix, cli))
	})
}
