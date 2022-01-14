package redis

import (
	"context"
	"sync"
	"time"

	instance "github.com/admiralbulldogtv/yappercontroller/src/instances"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type redisInstance struct {
	c       *redis.Client
	p       *redis.PubSub
	subs    map[string][]*redisSub
	subsMtx sync.Mutex
}

type SetupOptions struct {
	Username   string
	Password   string
	MasterName string
	Database   int

	Addresses []string
	Sentinel  bool
}

func NewInstance(ctx context.Context, opts SetupOptions) (instance.Redis, error) {
	if len(opts.Addresses) == 0 {
		logrus.Fatal("you must provide at least one redis address")
	}

	var rc *redis.Client

	if opts.Sentinel {
		rc = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       opts.MasterName,
			SentinelAddrs:    opts.Addresses,
			SentinelUsername: opts.Username,
			SentinelPassword: opts.Password,
			Username:         opts.Username,
			Password:         opts.Password,
			DB:               opts.Database,
		})
	} else {
		rc = redis.NewClient(&redis.Options{
			Addr:     opts.Addresses[0],
			Username: opts.Username,
			Password: opts.Password,
			DB:       opts.Database,
		})
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	i := &redisInstance{
		c:    rc,
		p:    rc.Subscribe(ctx),
		subs: make(map[string][]*redisSub),
	}

	if err := i.Ping(ctx); err != nil {
		_ = i.c.Close()
		return nil, err
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				logrus.WithField("err", err).Fatal("panic in subs")
			}
		}()
		ch := i.p.Channel()
		var msg *redis.Message
		for msg = range ch {
			payload := msg.Payload
			i.subsMtx.Lock()
			for _, s := range i.subs[msg.Channel] {
				go func(s *redisSub) {
					defer func() {
						if err := recover(); err != nil {
							logrus.WithField("err", err).Error("panic in subs")
						}
					}()
					s.ch <- payload
				}(s)
			}
			i.subsMtx.Unlock()
		}
	}()

	return i, nil
}

func (i *redisInstance) Ping(ctx context.Context) error {
	return i.c.Ping(ctx).Err()
}

func (i *redisInstance) SAdd(ctx context.Context, key string, values ...interface{}) error {
	return i.c.SAdd(ctx, key, values...).Err()
}

func (i *redisInstance) Set(ctx context.Context, key string, value string, expiry time.Duration) error {
	return i.c.Set(ctx, key, value, expiry).Err()
}

func (i *redisInstance) Get(ctx context.Context, key string) (string, error) {
	return i.c.Get(ctx, key).Result()
}
