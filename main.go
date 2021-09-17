package main

import (
	"context"

	"github.com/admiralbulldogtv/yappercontroller/src/configure"
	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/admiralbulldogtv/yappercontroller/src/manager"
	"github.com/admiralbulldogtv/yappercontroller/src/mongo"
	"github.com/admiralbulldogtv/yappercontroller/src/redis"
	"github.com/admiralbulldogtv/yappercontroller/src/tts"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := global.WithCancel(configure.Init(context.Background()))
	defer cancel()

	mongoInst, err := mongo.NewInstance(ctx, ctx.Config().MongoURI, ctx.Config().MongoDB)
	if err != nil {
		logrus.WithError(err).Fatal("failed to start mongo")
	}

	redisInst, err := redis.NewInstance(ctx, ctx.Config().RedisURI)
	if err != nil {
		logrus.WithError(err).Fatal("failed to start redis")
	}

	ttsInst, err := tts.NewInstance(ctx, ctx.Config().RedisTaskSetKey, ctx.Config().RedisOutputEvent)
	if err != nil {
		logrus.WithError(err).Fatal("failed to start tts")
	}

	ctx.SetMongoInstance(mongoInst)
	ctx.SetRedisInstance(redisInst)
	ctx.SetTtsInstance(ttsInst)

	done := manager.New(ctx)

	<-done
}
