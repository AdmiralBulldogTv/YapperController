package v1

import (
	"fmt"
	"strconv"

	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/admiralbulldogtv/yappercontroller/src/utils"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Wav(ctx global.Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		logrus.Info(c.Params("id"))
		id, err := primitive.ObjectIDFromHex(c.Params("id"))
		if err != nil {
			return c.SendStatus(404)
		}

		r := ctx.GetRedisInstance()
		result, err := r.Get(ctx, fmt.Sprintf("generated:tts:%s", id.Hex()))
		if err != nil {
			if err == redis.Nil {
				return c.SendStatus(404)
			}
			logrus.WithError(err).Error("failed to get tts from redis")
			return c.SendStatus(500)
		}

		data := utils.S2B(result)

		c.Set("Content-Type", "audio/wav")
		c.Set("Content-Length", strconv.Itoa(len(data)))

		return c.Status(200).Send(data)
	}
}
