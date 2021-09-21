package manager

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/alerts"
	"github.com/admiralbulldogtv/yappercontroller/src/datastructures"
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

var bitsRe = regexp.MustCompile(`\b(?:Cheer|BibleThump|cheerwhal|Corgo|uni|ShowLove|Party|SeemsGood|Pride|Kappa|FrankerZ|HeyGuys|DansGame|EleGiggle|TriHard|Kreygasm|4Head|SwiftRage|NotLikeThis|FailFish|VoHiYo|PJSalt|MrDestructoid|bday|RIPCheer|Shamrock)\d+\b`)

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

type alertHelper struct {
	Type string
	Name string
}

func (a alertHelper) Parse() (string, string) {
	switch a.Type {
	case "cheer":
		return alerts.CheerAlerts[a.Name+".gif"].ToName(), alerts.CheerAlerts[a.Name+".wav"].ToName()
	case "donation":
		return alerts.DonationAlerts[a.Name+".gif"].ToName(), alerts.DonationAlerts[a.Name+".wav"].ToName()
	case "subscriber":
		return alerts.SubscriberAlerts[a.Name+".gif"].ToName(), alerts.SubscriberAlerts[a.Name+".wav"].ToName()
	default:
		return "", ""
	}
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
				alert := alertHelper{}
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

				var (
					message      string
					alertText    string
					alertSubText string
				)

				switch evnt {
				case streamelements.EventListenerCheer:
					alert.Type = "cheer"
					alert.Name = "CheerDefault"

					data := streamelements.Cheer{}
					if err := json.Unmarshal(payload, &data); err != nil {
						logrus.WithError(err).Error("failed to parse event")
						continue event
					}
					message = data.Message
					alertSubText = data.Message
					alertText = fmt.Sprintf("~%s cheered ~%d bits", data.DisplayName, data.Amount)

					// filter bit emotes
					message = bitsRe.ReplaceAllString(message, "")

					amount := data.Amount / 100
					if amount < 3 {
						continue event
					}

					defaultVoice = textparser.VoicesMap["bull"]

					validVoices = append(validVoices,
						textparser.VoicesMap["bull"],
						textparser.VoicesMap["arno"],
						textparser.VoicesMap["krab"],
						textparser.VoicesMap["obama"],
					)

					if amount >= 5 {
						alert.Name = "Cheer500"
					}

					if amount >= 6 {
						validVoices = append(validVoices,
							textparser.VoicesMap["lac"],
							textparser.VoicesMap["glad"],
						)
					}

					if amount >= 10 {
						defaultVoice = textparser.VoicesMap["pooh"]
						validVoices = append(validVoices,
							textparser.VoicesMap["rae"],
							textparser.VoicesMap["pooh"],
						)

						alert.Name = "Cheer1000"
					}

					if amount >= 100 {
						alert.Name = "Cheer10000"
					}

					if amount >= 1000 {
						alert.Name = "Cheer100000"
					}
				case streamelements.EventListenerDonation:
					data := streamelements.Donation{}
					if err := json.Unmarshal(payload, &data); err != nil {
						logrus.WithError(err).Error("failed to parse event")
						continue event
					}
					message = data.Message
					alertSubText = data.Message
					alertText = fmt.Sprintf("~%s donated ~€%.2f", data.Name, data.Amount)

					alert.Type = "donation"
					alert.Name = "DonationDefault"

					defaultVoice = textparser.VoicesMap["bull"]

					validVoices = append(validVoices,
						textparser.VoicesMap["bull"],
						textparser.VoicesMap["arno"],
						textparser.VoicesMap["krab"],
						textparser.VoicesMap["obama"],
					)

					if data.Amount < 3 {
						continue event
					}

					if data.Amount == 4.2 {
						alert.Name = "Donation420"
					}

					if data.Amount >= 6 {
						validVoices = append(validVoices,
							textparser.VoicesMap["lac"],
							textparser.VoicesMap["glad"],
						)
					}

					if data.Amount >= 10 {
						defaultVoice = textparser.VoicesMap["pooh"]
						validVoices = append(validVoices,
							textparser.VoicesMap["rae"],
							textparser.VoicesMap["pooh"],
						)

						alert.Name = "Donation10"
					}

					if data.Amount >= 50 {
						alert.Name = "Donation50"
					}
				case streamelements.EventListenerSubscription:
					data := streamelements.Subscription{}
					if err := json.Unmarshal(payload, &data); err != nil {
						logrus.WithError(err).Error("failed to parse event")
						continue event
					}

					alert.Type = "subscriber"
					alert.Name = "Subscriber1"
					message = data.Message

					alertSubText = data.Message

					// ignore gifted subs.
					if data.Gifted {
						alert.Name = "SubscriberGift"
						message = ""
						alertText = fmt.Sprintf("~%s gifted a sub to ~%s", data.Sender, data.Name)
					} else if data.BulkGifted {
						alert.Name = "SubscriberGift"
						if data.Amount >= 5 {
							alert.Name = "SubscriberGift5"
						}
						if data.Amount >= 25 {
							alert.Name = "SubscriberGift25"
						}
						if data.Amount >= 95 {
							alert.Name = "SubscriberGift95"
						}
						alertText = fmt.Sprintf("~%s gifted ~%d subs", data.Sender, data.Amount)
					} else {
						alertText = fmt.Sprintf("~%s subscribed for ~%d months", data.Name, data.Amount)
						defaultVoice = textparser.VoicesMap["obama"]
						validVoices = append(validVoices, textparser.VoicesMap["bull"], textparser.VoicesMap["obama"])

						// voice calculation
						if data.Amount == 1 {
							alertText = fmt.Sprintf("~%s just subscribed", data.Name)
						}

						if data.Amount >= 2 {
							alert.Name = "Subscriber2"
						}

						if data.Amount >= 3 {
							alert.Name = "Subscriber3"
						}

						if data.Amount >= 6 {
							validVoices = append(validVoices, textparser.VoicesMap["obama"])

							alert.Name = "Subscriber6"
						}

						if data.Amount >= 9 {
							alert.Name = "Subscriber9"
						}

						if data.Amount >= 10 {
							validVoices = append(validVoices, textparser.VoicesMap["arno"])
						}

						if data.Amount >= 12 {
							alert.Name = "Subscriber12"
						}

						if data.Amount >= 13 {
							validVoices = append(validVoices, textparser.VoicesMap["lac"])
						}

						if data.Amount >= 18 {
							alert.Name = "Subscriber18"
						}

						if data.Amount >= 21 {
							validVoices = append(validVoices, textparser.VoicesMap["krab"])
						}

						if data.Amount >= 22 {
							validVoices = append(validVoices, textparser.VoicesMap["glad"])
						}

						if data.Amount >= 24 {
							alert.Name = "Subscriber24"
						}

						if data.Amount >= 30 {
							alert.Name = "Subscriber30"
						}

						if data.Amount >= 36 {
							alert.Name = "Subscriber36"
						}

						if data.Amount >= 41 {
							validVoices = append(validVoices, textparser.VoicesMap["rae"])
						}

						if data.Amount >= 42 {
							alert.Name = "Subscriber42"
						}

						if data.Amount >= 48 {
							alert.Name = "Subscriber48"
						}

						if data.Amount >= 54 {
							alert.Name = "Subscriber54"
						}

						if data.Amount >= 56 {
							validVoices = append(validVoices, textparser.VoicesMap["pooh"])
						}

						if data.Amount >= 60 {
							alert.Name = "Subscriber60"
						}

						if data.Tier == "2000" {
							alert.Name = "SubscriberSuper"
						}

						if data.Tier == "3000" {
							alert.Name = "SubscriberMega"
						}
					}

				default:
					continue event
				}

				logrus.Infof("generating tts from request %s", evnt)
				message = strings.TrimSpace(html.UnescapeString(message))
				alt := datastructures.SseEventTtsAlert{}
				image, audio := alert.Parse()
				alt.Audio = audio
				alt.Image = image
				alt.Text = alertText
				alt.SubText = strings.TrimSpace(html.UnescapeString(alertSubText))
				alt.Type = alert.Type
				go func(message string, alert datastructures.SseEventTtsAlert) {
					channelId, _ := primitive.ObjectIDFromHex(gCtx.Config().TtsChannelID)
					var id *primitive.ObjectID
					if message != "" {
						idt := primitive.NewObjectIDFromTimestamp(time.Now())
						id = &idt
					}
					if err := gCtx.GetTtsInstance().Generate(gCtx, message, id, channelId, defaultVoice, validVoices, 5, &alert); err != nil {
						logrus.WithError(err).Error("failed to generate tts")
					}
					logrus.Info("generated tts")
				}(message, alt)
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
