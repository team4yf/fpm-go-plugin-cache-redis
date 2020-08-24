package main

import (
	"time"

	"github.com/team4yf/yf-fpm-server-go/fpm"

	_ "github.com/team4yf/fpm-go-plugin-cache-redis/plugin"
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
	app.Run()

}
