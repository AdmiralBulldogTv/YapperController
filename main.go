package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/troydota/tts-textparser/src/configure"
	"github.com/troydota/tts-textparser/src/global"
	"github.com/troydota/tts-textparser/src/manager"
	"github.com/troydota/tts-textparser/src/mongo"
	"github.com/troydota/tts-textparser/src/redis"
	"github.com/troydota/tts-textparser/src/tts"
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
