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
	ctx, cancel := global.WithCancel(global.NewCtx(context.Background(), configure.New()))
	defer cancel()

	mongoInst, err := mongo.NewInstance(ctx, ctx.Config().Mongo.URI, ctx.Config().Mongo.Database)
	if err != nil {
		logrus.WithError(err).Fatal("failed to start mongo")
	}

	redisInst, err := redis.NewInstance(ctx, redis.SetupOptions{
		Username:   ctx.Config().Redis.Username,
		Password:   ctx.Config().Redis.Password,
		MasterName: ctx.Config().Redis.MasterName,
		Database:   ctx.Config().Redis.Database,
		Addresses:  ctx.Config().Redis.Addresses,
		Sentinel:   ctx.Config().Redis.Sentinel,
	})
	if err != nil {
		logrus.WithError(err).Fatal("failed to start redis")
	}

	ttsInst, err := tts.NewInstance(ctx, ctx.Config().Redis.TaskSetKey, ctx.Config().Redis.OutputEvent)
	if err != nil {
		logrus.WithError(err).Fatal("failed to start tts")
	}

	ctx.Inst().Mongo = mongoInst
	ctx.Inst().Redis = redisInst
	ctx.Inst().TTS = ttsInst

	done := manager.New(ctx)

	<-done
}
