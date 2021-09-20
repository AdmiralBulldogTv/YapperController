package v1

import (
	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Api(ctx global.Context, app fiber.Router) {
	app.Get("/sse/:token", SSE(ctx))

	app.Get("/wav/:id.wav", Wav(ctx))

	Twitch(ctx, app.Group("/twitch"))

	Alerts(app.Group("/alerts"))
}
