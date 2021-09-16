package manager

import (
	"fmt"
	"html"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/troydota/tts-textparser/src/global"
	"github.com/troydota/tts-textparser/src/server"
	"github.com/troydota/tts-textparser/src/streamelements"
	"github.com/troydota/tts-textparser/src/twitch"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func New(ctx global.Context) <-chan struct{} {
	done := make(chan struct{})

	manager := &Manager{}

	if ctx.Config().StreamElementsEnabled {
		manager.se = streamelements.NewClient()
		if err := manager.handleSe(ctx); err != nil {
			logrus.WithError(err).Fatal("streamelements failed")
		}
		logrus.Info("streamelements started")
	}

	serverDone := server.New(ctx)

	go func() {
		<-ctx.Done()
		<-serverDone
		close(done)
	}()

	_, err := twitch.NewClient(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("twitch failed")
	}

	return done
}

type Manager struct {
	se streamelements.Client
}

func (m *Manager) handleSe(gCtx global.Context) error {
	ctx, cancel := global.WithTimeout(gCtx, time.Second*10)
	defer cancel()

	che := make(chan error)
	defer close(che)

	go func() {
		ch := m.se.Events()
	event:
		for event := range ch {
			switch event.Name {
			case "connect":
				logrus.Info("streamelements connected")
				if err := m.se.Auth(gCtx.Config().StreamElementsAuthMethod, gCtx.Config().StreamElementsAuthToken); err != nil {
					panic(err)
				}
			case "disconnect":
				logrus.Warn("streamelements disconnected")
				time.Sleep(time.Second)
				if err := m.handleSe(gCtx); err != nil {
					panic(err)
				}
			case "authenticated":
				che <- nil
			case "unauthorized":
				che <- fmt.Errorf("%s", event.Payload)
			case "event:test":
				// donation/sub/cheer
				se := streamelements.EventPayload{}
				if err := json.Unmarshal(event.Payload, &se); err != nil {
					logrus.WithError(err).Error("failed to parse event")
					continue
				}
				logrus.Infof("generating tts from request %s", se.Listener)
				var message string
				switch se.Listener {
				case streamelements.EventListenerCheer:
					data := streamelements.Cheer{}
					if err := json.Unmarshal(se.Event, &data); err != nil {
						logrus.WithError(err).Error("failed to parse event")
						continue event
					}
					message = data.Message
				case streamelements.EventListenerDonation:
					data := streamelements.Donation{}
					if err := json.Unmarshal(se.Event, &data); err != nil {
						logrus.WithError(err).Error("failed to parse event")
						continue event
					}
					message = data.Message
				case streamelements.EventListenerSubscription:
					data := streamelements.Subscription{}
					if err := json.Unmarshal(se.Event, &data); err != nil {
						logrus.WithError(err).Error("failed to parse event")
						continue event
					}
					message = data.Message
				}
				message = strings.TrimSpace(html.UnescapeString(message))
				if message != "" {
					go func() {
						channelId, _ := primitive.ObjectIDFromHex(gCtx.Config().TtsChannelID)
						_, err := gCtx.GetTtsInstance().Generate(gCtx, message, primitive.NewObjectIDFromTimestamp(time.Now()), channelId)
						if err != nil {
							logrus.WithError(err).Error("failed to generate tts")
						}
						logrus.Info("generated tts")
					}()
				}
			}
		}
	}()

	if err := m.se.Connect(ctx.Config().StreamElementsWssUrl); err != nil {
		return err
	}

	select {
	case err := <-che:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
