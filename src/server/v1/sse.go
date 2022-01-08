package v1

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SSE(ctx global.Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		tkn, err := primitive.ObjectIDFromHex(c.Params("token"))
		if err != nil {
			return c.SendStatus(401)
		}
		mgo := ctx.Inst().Mongo
		overlay, err := mgo.FetchOverlay(c.Context(), tkn)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.SendStatus(401)
			}
			logrus.WithError(err).Error("failed to fetch overlay")
			return c.SendStatus(500)
		}

		reqCtx := c.Context()

		localCtx, cancel := context.WithCancel(context.Background())
		subCh := make(chan string)

		redis := ctx.Inst().Redis
		redis.Subscribe(localCtx, subCh, fmt.Sprintf("overlay:events:%s", overlay.ChannelID.Hex()))

		go func() {
			defer func() {
				cancel()
				close(subCh)
			}()
			select {
			case <-ctx.Done():
			case <-localCtx.Done():
			}
		}()

		reqCtx.SetContentType("text/event-stream")
		reqCtx.Response.Header.Set("Cache-Control", "no-cache")
		reqCtx.Response.Header.Set("Connection", "keep-alive")
		reqCtx.Response.Header.Set("Transfer-Encoding", "chunked")
		reqCtx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		reqCtx.Response.Header.Set("Access-Control-Allow-Headers", "Cache-Control")
		reqCtx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
		reqCtx.Response.Header.Set("X-Accel-Buffering", "no")

		reqCtx.SetBodyStreamWriter(func(w *bufio.Writer) {
			tick := time.NewTicker(time.Second * 30)
			defer func() {
				defer cancel()
				_ = w.Flush()
				tick.Stop()
			}()
			var (
				msg string
				err error
			)
			if _, err = w.WriteString("event: ready\ndata: tts-event-sub.v1\n\n"); err != nil {
				return
			}
			if err = w.Flush(); err != nil {
				return
			}
			for {
				select {
				case <-localCtx.Done():
					return
				case <-tick.C:
					if _, err = w.WriteString("event: heartbeat\n"); err != nil {
						return
					}
					if _, err = w.WriteString("data: {}\n\n"); err != nil {
						return
					}
					if err = w.Flush(); err != nil {
						return
					}
				case msg = <-subCh:
					if _, err = w.WriteString("event: action\n"); err != nil {
						return
					}
					if _, err = w.WriteString("data: "); err != nil {
						return
					}
					if _, err = w.WriteString(msg); err != nil {
						return
					}
					// Write a `\n\n` bytes to signify end of a message to signify end of event.
					if _, err = w.WriteString("\n\n"); err != nil {
						return
					}
					if err = w.Flush(); err != nil {
						return
					}
				}
			}
		})

		return nil
	}
}
