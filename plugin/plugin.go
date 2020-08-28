//Package plugin the plugin body
package plugin

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/team4yf/fpm-go-plugin-cache-redis/rds"
	"github.com/team4yf/yf-fpm-server-go/fpm"
)

//Options the redis config
type Options struct {
	Prefix string
	Addr   string
	Passwd string
	DB     int
	Pool   int
}

func contains(arr []string, ele string) bool {
	for _, v := range arr {
		if ele == v {
			return true
		}
	}
	return false
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
		if _, err := cli.Ping(rds.TimeoutCtx).Result(); err != nil {
			app.Logger.Errorf("Redis Cant Connect! Cause: %v", err)
			panic(err)
		}

		cacher := rds.NewRedisCache(options.Prefix, cli)
		locker := rds.NewRedisLocker(cli)

		app.SetCacher(cacher)
		app.SetDistributeLocker(locker)

		// Important:
		// 1. the subscribe consumer will shut when the server shuted
		//    we should autorun the subscriber
		//    so, we should save all the events into the redis;
		//    and fetch them when startup, and do the subscribe.
		// 2. subTopics include all the topics what have been subscribed.

		var sub *redis.PubSub
		storedKey := "__subscribe__topics"
		getSubscribeTopics := func() (subTopics []string) {
			subTopics = make([]string, 0)
			if ok, _ := cacher.IsSet(storedKey); !ok {
				return
			}
			_, err := cacher.GetObject(storedKey, &subTopics)
			if err != nil {
				app.Logger.Errorf("Redis do GetObject func error: %v", err)
				panic(err)
			}
			return
		}

		subscribe := func(events interface{}) {
			locker.GetLock("sync_to_subscribe", 10*time.Second)
			defer locker.ReleaseLock("sync_to_subscribe")
			subTopics := getSubscribeTopics()
			topics := make([]string, 0)
			switch events.(type) {
			case string:
				topics = append(topics, events.(string))
			case []string:
				topics = events.([]string)
			case []interface{}:
				for _, t := range events.([]interface{}) {
					topics = append(topics, t.(string))
				}
			}
			// exclude the topic if contains in the global subTopics
			//
			finalTopics := make([]string, 0)
			for _, t := range topics {
				if !contains(subTopics, t) {
					finalTopics = append(finalTopics, t)
				}
			}
			if len(finalTopics) < 1 {
				//Nothing to do
				return
			}
			sub = cli.Subscribe(rds.TimeoutCtx, finalTopics...)
			// append into the subscribe topics.
			subTopics = append(subTopics, finalTopics...)

			if err := cacher.SetObject(storedKey, &subTopics, 0); err != nil {
				app.Logger.Errorf("Redis do SetObject error: %v", err)
			}

			go func() {
				defer sub.Close()
				for {
					msg, _ := sub.ReceiveMessage(rds.TimeoutCtx)
					app.Publish("#redis/receive", map[string]interface{}{
						"topic":   msg.Channel,
						"payload": msg.Payload,
					})
				}
			}()
		}
		unsubscribe := func(events interface{}) {
			locker.GetLock("sync_to_unsubscribe", 10*time.Second)
			defer locker.ReleaseLock("sync_to_unsubscribe")
			subTopics := getSubscribeTopics()
			topics := make([]string, 0)
			switch events.(type) {
			case string:
				topics = append(topics, events.(string))
			case []string:
				topics = events.([]string)
			case []interface{}:
				for _, t := range events.([]interface{}) {
					topics = append(topics, t.(string))
				}
			}

			finalTopics := make([]string, 0)
			for _, t := range subTopics {
				// append into the final topics if not contains in the unsubscribe topics
				if !contains(topics, t) {
					finalTopics = append(finalTopics, t)
				} else {
					//Do unsubscribe
					if sub != nil {
						sub.Unsubscribe(rds.TimeoutCtx, subTopics...)
					}
					subscribe(finalTopics)
				}
			}
			if err := cacher.SetObject(storedKey, &finalTopics, 0); err != nil {
				app.Logger.Errorf("Redis do SetObject error: %v", err)
			}
		}

		// load the topics saved in the redis, fetch them and do subscribe
		ok, err := cacher.IsSet(storedKey)
		if err != nil {
			app.Logger.Errorf("Redis do isset func error: %v", err)
			panic(err)
		}
		if ok {
			topics := make([]string, 0)
			_, err := cacher.GetObject(storedKey, &topics)
			if err != nil {
				app.Logger.Errorf("Redis do GetObject func error: %v", err)
				panic(err)
			}
			app.Logger.Debugf("fetch __subscribe__topics from redis: %v", topics)
			subscribe(topics)
		}

		bizModule := make(fpm.BizModule, 0)

		// Warnning:
		// 1. the subscribe method can be invoke multi times.
		//    it will create a sub consumer once when the method called once.
		//    should make sure the topics are seted, there could not contain same key.
		bizModule["subscribe"] = func(param *fpm.BizParam) (data interface{}, err error) {
			topic := (*param)["topic"]
			subscribe(topic)
			data = 1
			return
		}
		bizModule["unsubscribe"] = func(param *fpm.BizParam) (data interface{}, err error) {
			topic := (*param)["topic"]
			unsubscribe(topic)
			data = 1
			return
		}

		bizModule["publish"] = func(param *fpm.BizParam) (data interface{}, err error) {
			topic := (*param)["topic"].(string)
			payload := (*param)["payload"]
			err = cli.Publish(rds.TimeoutCtx, topic, payload).Err()
			data = 1
			return
		}
		app.AddBizModule("redis", &bizModule)
	})
}
