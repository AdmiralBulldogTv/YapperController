package health

import (
	"context"
	"sync"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/gofiber/fiber/v2"

	log "github.com/sirupsen/logrus"
)

func Health(ctx global.Context, app fiber.Router) {
	downedServices := map[string]bool{
		"redis": false,
		"mongo": false,
	}

	redis := ctx.GetRedisInstance()
	mongo := ctx.GetMongoInstance()

	mtx := sync.Mutex{}
	app.Get("/", func(c *fiber.Ctx) error {
		mtx.Lock()
		defer mtx.Unlock()

		isDown := false

		redisCtx, cancel := context.WithTimeout(c.Context(), time.Second*10)
		defer cancel()
		// CHECK REDIS
		if err := redis.Ping(redisCtx); err != nil {
			log.Error("health, REDIS IS DOWN")
			isDown = true
			if downedServices["redis"] {
				downedServices["redis"] = true
			}
		} else {
			if downedServices["redis"] {
				downedServices["redis"] = false
			}
		}

		// CHECK MONGO
		mongoCtx, cancel := context.WithTimeout(c.Context(), time.Second*10)
		defer cancel()
		if err := mongo.Ping(mongoCtx); err != nil {
			log.Error("health, MONGO IS DOWN")
			isDown = true
			if !downedServices["mongo"] {
				downedServices["mongo"] = true
			}
		} else {
			if downedServices["mongo"] {
				downedServices["mongo"] = false
			}
		}

		if isDown {
			return c.SendStatus(503)
		}

		return c.Status(200).SendString("OK")
	})

}
