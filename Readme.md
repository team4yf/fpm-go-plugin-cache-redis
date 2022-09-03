## redis plugin 

> A plugin of redis client, for fpm-server-go.

### Install

`$ go get -u github.com/team4yf/fpm-go-plugin-cache-redis`

### Config

```yaml
redis:
  prefix: test
  addr: localhost:6379
  passwd:
  db: 1
  pool: 10
```

### Usage

0. import

```golang
import (
    //...

    //import the fpm server core
    "github.com/team4yf/yf-fpm-server-go/fpm"

    //import the plugin
	_ "github.com/team4yf/fpm-go-plugin-cache-redis/plugin"
    //...
)
```

1. use it as cacher and locker

```golang
func main(){
    //new a fpm instance
    app := fpm.New()
	app.Init()

    //get the redis api
	c, exists := app.GetCacher()
	if exists {
		if err := c.SetInt("test", 1024, 60*time.Second); err != nil {
			app.Logger.Errorf("setInt error: %v", err)

		} else {
			data, _ := c.GetInt("test")
			app.Logger.Debugf("getInt data: %v", data)
		}

    }
    //get the redis locker
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

```

2. use it as pub/sub

```golang
func main(){
    //new a fpm instance
    app := fpm.New()
	app.Init()

    //pub message, payload support interface{}, it will be serialized to []byte
	app.Execute("redis.publish", &fpm.BizParam{
		"topic":   "sometopic",
		"payload": `{"test":1}`,
    })

    //sub topics
    app.Execute("redis.subscribe", &fpm.BizParam{
		"topics": []string{"sometopic1", "sometopic2"},
    })
    //catch the message
    app.Subscribe("#redis/receive", func(topic string, data interface{}) {
		//data 通常是 byte[] 类型，可以转成 string 或者 map
		body := data.(map[string]interface{})
		t := body["topic"].(string)
		p := body["payload"].([]byte)
		log.Debugf("topic: %s, payload: %+v", t, (string)(p))
	})
	app.Run()
}

```

3. Get original redis client

```golang
import "github.com/team4yf/fpm-go-plugin-cache-redis/plugin"

cli := plugin.GetClient()
// TODO: do something here

```