package v1

import (
	"strconv"
	"strings"

	"github.com/admiralbulldogtv/yappercontroller/src/alerts"
	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/gofiber/fiber/v2"
)

func Alerts(ctx global.Context, app fiber.Router) {
	lookUpFn := func(mp map[string]alerts.Alert) func(c *fiber.Ctx) error {
		return func(c *fiber.Ctx) error {
			file := c.Params("*")
			ctype := ""
			if strings.HasSuffix(file, ".wav") {
				ctype = "audio/wav"
			} else if strings.HasSuffix(file, ".gif") {
				ctype = "	image/gif"
			} else {
				return c.SendStatus(404)
			}

			idx := strings.LastIndexByte(file, '.')
			suffix := file[idx:]
			file = file[:idx]
			idx = strings.LastIndexByte(file, '.')
			sum := file[idx+1:]
			file = file[:idx]

			if v, ok := mp[file+suffix]; ok {
				if v.CheckSum == sum {
					c.Set("Content-Type", ctype)
					c.Set("Content-Length", strconv.Itoa(len(v.Data)))
					return c.Status(200).Send(v.Data)
				}
			}

			return c.SendStatus(404)
		}
	}

	app.Get("/cheer/*", lookUpFn(alerts.CheerAlerts))
	app.Get("/donation/*", lookUpFn(alerts.DonationAlerts))
	app.Get("/subscriber/*", lookUpFn(alerts.SubscriberAlerts))
}
