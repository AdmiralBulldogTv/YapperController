package redis

import (
	"context"

	"github.com/sirupsen/logrus"
)

type redisSub struct {
	ch chan string
}

// Publish to a redis channel
func (inst *redisInstance) Publish(ctx context.Context, channel string, data string) error {
	return inst.c.Publish(ctx, channel, data).Err()
}

// Subscribe to a channel on Redis
func (inst *redisInstance) Subscribe(ctx context.Context, ch chan string, subscribeTo ...string) {
	inst.subsMtx.Lock()
	defer inst.subsMtx.Unlock()
	localSub := &redisSub{ch}
	for _, e := range subscribeTo {
		if _, ok := inst.subs[e]; !ok {
			if err := inst.p.Subscribe(ctx, e); err != nil {
				panic(err)
			}
		}
		inst.subs[e] = append(inst.subs[e], localSub)
	}

	go func() {
		<-ctx.Done()
		inst.subsMtx.Lock()
		defer inst.subsMtx.Unlock()
		for _, e := range subscribeTo {
			for i, v := range inst.subs[e] {
				if v == localSub {
					if i != len(inst.subs[e])-1 {
						inst.subs[e][i] = inst.subs[e][len(inst.subs[e])-1]
					}
					inst.subs[e] = inst.subs[e][:len(inst.subs[e])-1]
					if len(inst.subs[e]) == 0 {
						delete(inst.subs, e)
						if err := inst.p.Unsubscribe(context.Background(), e); err != nil {
							logrus.WithError(err).Error("failed to unsubscribe")
						}
					}
					break
				}
			}
		}
	}()
}
