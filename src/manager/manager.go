package manager

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/admiralbulldogtv/yappercontroller/src/server"
	"github.com/admiralbulldogtv/yappercontroller/src/streamelements"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
	"github.com/admiralbulldogtv/yappercontroller/src/twitch"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
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
			case "event:update", "event:test":
				defaultVoice := textparser.VoicesMap["ann1"]
				validVoices := []parts.Voice{
					textparser.VoicesMap["ann1"],
					textparser.VoicesMap["narr1"],
					textparser.VoicesMap["narr2"],
					textparser.VoicesMap["narr3"],
				}

				var evnt string
				var payload jsoniter.RawMessage

				if event.Name == "event:update" {
					// donation/sub/cheer
					se := streamelements.EventUpdatePayload{}
					if err := json.Unmarshal(event.Payload, &se); err != nil {
						logrus.WithError(err).Error("failed to parse event")
						continue
					}
					evnt = se.Name
					payload = se.Data
				} else {
					// donation/sub/cheer
					se := streamelements.EventTestPayload{}
					if err := json.Unmarshal(event.Payload, &se); err != nil {
						logrus.WithError(err).Error("failed to parse event")
						continue
					}
					evnt = se.Listener
					payload = se.Event
				}

				var message string
				switch evnt {
				case streamelements.EventListenerCheer:
					data := streamelements.Cheer{}
					if err := json.Unmarshal(payload, &data); err != nil {
						logrus.WithError(err).Error("failed to parse event")
						continue event
					}
					message = data.Message
					amount := data.Amount / 100

					defaultVoice = textparser.VoicesMap["bull"]

					validVoices = append(validVoices,
						textparser.VoicesMap["bull"],
						textparser.VoicesMap["arno"],
					)

					if amount >= 6 {
						validVoices = append(validVoices,
							textparser.VoicesMap["lac"],
							textparser.VoicesMap["glad"],
						)
					}

					if amount >= 10 {
						validVoices = append(validVoices,
							textparser.VoicesMap["rae"],
							textparser.VoicesMap["pooh"],
							textparser.VoicesMap["krab"],
						)
					}
				case streamelements.EventListenerDonation:
					data := streamelements.Donation{}
					if err := json.Unmarshal(payload, &data); err != nil {
						logrus.WithError(err).Error("failed to parse event")
						continue event
					}
					message = data.Message

					defaultVoice = textparser.VoicesMap["bull"]

					validVoices = append(validVoices,
						textparser.VoicesMap["bull"],
						textparser.VoicesMap["arno"],
					)

					if data.Amount >= 6 {
						validVoices = append(validVoices,
							textparser.VoicesMap["lac"],
							textparser.VoicesMap["glad"],
						)
					}

					if data.Amount >= 10 {
						validVoices = append(validVoices,
							textparser.VoicesMap["rae"],
							textparser.VoicesMap["pooh"],
							textparser.VoicesMap["krab"],
						)
					}
				case streamelements.EventListenerSubscription:
					data := streamelements.Subscription{}
					if err := json.Unmarshal(payload, &data); err != nil {
						logrus.WithError(err).Error("failed to parse event")
						continue event
					}
					message = data.Message

					// voice calculation
					if data.Amount >= 1 {
						defaultVoice = textparser.VoicesMap["bull"]
						validVoices = append(validVoices, textparser.VoicesMap["bull"])
					}

					if data.Amount >= 6 {
						defaultVoice = textparser.VoicesMap["obama"]
						validVoices = append(validVoices, textparser.VoicesMap["obama"])
					}

					if data.Amount >= 10 {
						defaultVoice = textparser.VoicesMap["arno"]
						validVoices = append(validVoices, textparser.VoicesMap["arno"])
					}

					if data.Amount >= 13 {
						defaultVoice = textparser.VoicesMap["lac"]
						validVoices = append(validVoices, textparser.VoicesMap["lac"])
					}

					if data.Amount >= 21 {
						defaultVoice = textparser.VoicesMap["krab"]
						validVoices = append(validVoices, textparser.VoicesMap["krab"])
					}

					if data.Amount >= 22 {
						defaultVoice = textparser.VoicesMap["glad"]
						validVoices = append(validVoices, textparser.VoicesMap["glad"])
					}

					if data.Amount >= 30 {
						defaultVoice = textparser.VoicesMap["bull"]
					}

					if data.Amount >= 41 {
						defaultVoice = textparser.VoicesMap["rae"]
						validVoices = append(validVoices, textparser.VoicesMap["rae"])
					}

					if data.Amount >= 56 {
						defaultVoice = textparser.VoicesMap["pooh"]
						validVoices = append(validVoices, textparser.VoicesMap["pooh"])
					}

				default:
					continue event
				}

				logrus.Infof("generating tts from request %s", evnt)
				message = strings.TrimSpace(html.UnescapeString(message))
				if message != "" {
					go func() {
						channelId, _ := primitive.ObjectIDFromHex(gCtx.Config().TtsChannelID)
						_, err := gCtx.GetTtsInstance().Generate(gCtx, message, primitive.NewObjectIDFromTimestamp(time.Now()), channelId, defaultVoice, validVoices)
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
