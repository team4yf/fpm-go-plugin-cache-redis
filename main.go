package main

import (
	"context"
	"time"

	"github.com/team4yf/yf-fpm-server-go/fpm"

	"github.com/team4yf/fpm-go-plugin-cache-redis/plugin"
)

func main() {

	app := fpm.New()
	app.Init()

	c, exists := app.GetCacher()
	if exists {
		if err := c.SetInt("test", 1024, 60*time.Second); err != nil {
			app.Logger.Errorf("setInt error: %v", err)

		} else {
			data, _ := c.GetInt("test")
			app.Logger.Debugf("getInt data: %v", data)
		}

	}
	l, exists := app.GetDistributeLocker()
	if exists {
		if l.GetLock("a", 10*time.Second) {
			app.Logger.Debugf("getLocker")
			time.Sleep(3 * time.Second)
			if err := l.ReleaseLock("a"); err == nil {
				app.Logger.Debugf("releaseLocker")
			}
		}

	}

	app.Subscribe("#redis/receive", func(_ string, data interface{}) {
		app.Logger.Debugf("receive redis message %v", data)
	})
	app.Execute("redis.subscribe", &fpm.BizParam{
		"topic": "foo",
	}, nil)
	time.Sleep(5 * time.Second)
	app.Execute("redis.publish", &fpm.BizParam{
		"topic":   "foo",
		"payload": "bar",
	}, nil)

	cli := plugin.GetClient()
	if cmd := cli.LPush(context.Background(), "test", "abc", "bcd"); cmd.Err() != nil {
		panic(cmd.Err())
	}
	app.Run()

}
