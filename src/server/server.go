package server

import (
	"strings"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/admiralbulldogtv/yappercontroller/src/server/health"
	"github.com/admiralbulldogtv/yappercontroller/src/server/middleware"
	v1 "github.com/admiralbulldogtv/yappercontroller/src/server/v1"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sirupsen/logrus"
)

func New(ctx global.Context) <-chan struct{} {
	done := make(chan struct{})

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		DisableKeepalive:      true,
	})

	app.Use(middleware.Logger())
	app.Use(middleware.Xss())

	app.Use(cors.New(cors.Config{
		AllowOrigins: strings.Join(ctx.Config().Cors, " "),
	}))

	health.Health(ctx, app.Group("/health"))
	v1.Api(ctx, app.Group("/v1"))

	app.Use("/", func(c *fiber.Ctx) error {
		return c.SendStatus(404)
	})

	go func() {
		if err := app.Listen(ctx.Config().ApiBind); err != nil {
			logrus.WithError(err).Fatal("server failed")
		}
	}()

	go func() {
		<-ctx.Done()
		_ = app.Shutdown()
		close(done)
	}()

	logrus.Infof("Api started on %s", ctx.Config().ApiBind)

	return done
}
