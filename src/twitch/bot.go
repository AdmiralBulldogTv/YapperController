package twitch

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/datastructures"
	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser/parts"
	"github.com/davecgh/go-spew/spew"
	"github.com/gempir/go-twitch-irc/v2"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/go-multierror"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Client interface {
	SendMessage(channel string, message string) error
	SendWhisper(username string, message string) error
}

type TokenResp struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	Scope        []string  `json:"scope"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type twitchClient struct {
	cl *twitch.Client
}

func NewClient(ctx global.Context) (Client, error) {
	r := ctx.GetRedisInstance()
	ch := make(chan string)
	r.Subscribe(ctx, ch, "events:twitch:bot:login")

	data, err := r.Get(ctx, fmt.Sprintf("twitch:bot:%s", ctx.Config().TwitchBotID))
	if err != nil {
		if err == redis.Nil {
			for botID := range ch {
				if botID == ctx.Config().TwitchBotID {
					break
				}
			}
			data, err = r.Get(ctx, fmt.Sprintf("twitch:bot:%s", ctx.Config().TwitchBotID))
			if err != nil {
				return nil, err
			}
		}
	}

	token := TokenResp{}
	if err = json.UnmarshalFromString(data, &token); err != nil {
		return nil, err
	}

	if token.ExpiresAt.Before(time.Now()) {
		// refresh token
		token, err = RefreshToken(ctx, token.RefreshToken)
		if err != nil {
			return nil, err
		}

		data, _ = json.MarshalToString(token)
		if err = r.Set(ctx, fmt.Sprintf("twitch:bot:%s", ctx.Config().TwitchBotID), data, 0); err != nil {
			return nil, err
		}
	}

	client := &twitchClient{}
	client.cl = twitch.NewClient(ctx.Config().TwitchBotUsername, fmt.Sprintf("oauth:%s", token.AccessToken))

	client.cl.Join(ctx.Config().TwitchBotControlChannel, ctx.Config().TwitchStreamerChannel)
	client.cl.OnWhisperMessage(func(message twitch.WhisperMessage) {
		found := false
		for _, v := range ctx.Config().WhitelistedTwitchAccounts {
			if v == message.User.ID {
				found = true
				break
			}
		}
		if !found {
			return
		}

		msg := strings.TrimSpace(message.Message)
		channelID, _ := primitive.ObjectIDFromHex(ctx.Config().TtsChannelID)
		if strings.HasPrefix(msg, "!say ") {
			msg = strings.TrimPrefix(msg, "!say ")
			id := primitive.NewObjectIDFromTimestamp(time.Now())
			if err := ctx.GetTtsInstance().Generate(ctx, msg, &id, channelID, textparser.Voices[0], textparser.Voices, 30, nil); err != nil {
				err = multierror.Append(err, client.SendWhisper(message.User.Name, "failed to generate tts"))
				logrus.WithError(err).Error("failed to generate tts")
				return
			}
			_ = client.SendWhisper(message.User.Name, "generated tts")
		} else if msg == "!skip" {
			err := ctx.GetTtsInstance().Skip(ctx, channelID)
			if err != nil {
				err = multierror.Append(err, client.SendWhisper(message.User.Name, "failed to skip tts"))
				logrus.WithError(err).Error("failed to skip tts")
				return
			}
			_ = client.SendWhisper(message.User.Name, "skipped tts")
		} else if msg == "!reload" {
			err := ctx.GetTtsInstance().Reload(ctx, channelID)
			if err != nil {
				_ = client.SendWhisper(message.User.Name, "failed to reload overlay")
				logrus.WithError(err).Error("failed to reload overlay")
				return
			}
			_ = client.SendWhisper(message.User.Name, "reloaded overlay")
		}
	})

	client.cl.OnPrivateMessage(func(message twitch.PrivateMessage) {
		// ignore non control channel messages.
		if !strings.EqualFold(message.Channel, ctx.Config().TwitchBotControlChannel) {
			return
		}

		found := false
		for _, v := range ctx.Config().WhitelistedTwitchAccounts {
			if v == message.User.ID {
				found = true
				break
			}
		}
		if !found {
			return
		}

		msg := strings.TrimSpace(message.Message)
		channelID, _ := primitive.ObjectIDFromHex(ctx.Config().TtsChannelID)
		if strings.HasPrefix(msg, "!say ") {
			msg = strings.TrimPrefix(msg, "!say ")
			id := primitive.NewObjectIDFromTimestamp(time.Now())
			if err := ctx.GetTtsInstance().Generate(ctx, msg, &id, channelID, textparser.Voices[0], textparser.Voices, 30, nil); err != nil {
				err = multierror.Append(err, client.SendMessage(message.Channel, fmt.Sprintf("@%s, failed to generate tts", message.User.DisplayName)))
				logrus.WithError(err).Error("failed to generate tts")
				return
			}
			_ = client.SendMessage(message.Channel, fmt.Sprintf("@%s, generated tts", message.User.DisplayName))
		} else if msg == "!skip" {
			err := ctx.GetTtsInstance().Skip(ctx, channelID)
			if err != nil {
				_ = client.SendMessage(message.Channel, fmt.Sprintf("@%s, failed to skip tts", message.User.DisplayName))
				logrus.WithError(err).Error("failed to skip tts")
				return
			}
			_ = client.SendMessage(message.Channel, fmt.Sprintf("@%s, skipped tts", message.User.DisplayName))
		} else if msg == "!reload" {
			err := ctx.GetTtsInstance().Reload(ctx, channelID)
			if err != nil {
				_ = client.SendMessage(message.Channel, fmt.Sprintf("@%s, failed to reload overlay", message.User.DisplayName))
				logrus.WithError(err).Error("failed to reload overlay")
				return
			}
			_ = client.SendMessage(message.Channel, fmt.Sprintf("@%s, reloaded overlay", message.User.DisplayName))
		}
	})

	mtx := sync.Mutex{}
	// in theory this map will grow really really large.
	// thus every hour or so we need to go through the map and remove old ones.
	bulkGiftMap := map[string]time.Time{}
	go func() {
		tick := time.NewTicker(time.Hour)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				mtx.Lock()
				cutOff := time.Now().Add(-time.Hour)
				for k, v := range bulkGiftMap {
					if v.Before(cutOff) {
						delete(bulkGiftMap, k)
					}
				}
				mtx.Unlock()
			}
		}
	}()

	client.cl.OnUserNoticeMessage(func(message twitch.UserNoticeMessage) {
		mtx.Lock()
		defer mtx.Unlock()
		alert := datastructures.AlertHelper{}
		alertText := ""
		subText := ""
		spew.Dump(message.MsgParams)
		switch message.MsgID {
		case "submysterygift": // multi gift subs
			giftCount, err := strconv.Atoi(message.MsgParams["mass-gift-count"])
			if err != nil {
				logrus.WithError(err).Error("bad read from gift subs")
				return
			}
			if giftCount == 1 {
				// single gifts are handled by "subgift" event.
				return
			}
			// make sure the "subgift" event does not process these events.
			bulkGiftMap[message.MsgParams["origin-id"]] = time.Now()
			// we can now handle this a bulk gift sub.
			senderCount, err := strconv.Atoi(message.MsgParams["sender-count"])
			if err != nil {
				logrus.WithError(err).Error("bad read from gift subs")
				return
			}
			subPlan := message.MsgParams["sub-plan"]
			displayName := message.User.DisplayName

			alert.Name = "SubscriberGift"
			if giftCount >= 5 {
				alert.Name = "SubscriberGift5"
			}
			if giftCount >= 25 {
				alert.Name = "SubscriberGift25"
			}
			if giftCount >= 95 {
				alert.Name = "SubscriberGift95"
			}
			alertText = fmt.Sprintf("~%s gifted ~%d subs", displayName, giftCount)
			if subPlan != "1000" {
				alertText = fmt.Sprintf("~%s gifted ~%d tier %s subs", displayName, giftCount, string(subPlan[0]))
			}
			subText = fmt.Sprintf("they have gifted %d subs to the channel", senderCount)
		case "subgift":
			if !bulkGiftMap[message.MsgParams["origin-id"]].IsZero() {
				// has been already handled by the mystrygift event.
				return
			}
			// single giftsub.
			alert.Name = "SubscriberGift"
			displayName := message.User.DisplayName
			recipientDisplayName := message.MsgParams["recipient-display-name"]
			alertText = fmt.Sprintf("~%s gifted a sub to ~%s", displayName, recipientDisplayName)

			senderCount, err := strconv.Atoi(message.MsgParams["sender-count"])
			if err != nil {
				logrus.WithError(err).Error("bad read from gift subs")
				return
			}
			subText = fmt.Sprintf("they have gifted %d subs to the channel", senderCount)
			if senderCount == 1 {
				subText = "this is their first gifted sub"
			}
		default:
			// ignore all other twitch events.
			return
		}

		alt := datastructures.SseEventTtsAlert{}
		image, audio := alert.Parse()
		alt.Audio = audio
		alt.Image = image
		alt.Text = alertText
		alt.SubText = subText
		alt.Type = alert.Type
		go func(alert datastructures.SseEventTtsAlert) {
			channelId, _ := primitive.ObjectIDFromHex(ctx.Config().TtsChannelID)
			if err := ctx.GetTtsInstance().Generate(ctx, "", nil, channelId, parts.Voice{}, nil, 0, &alert); err != nil {
				logrus.WithError(err).Error("failed to generate tts")
			}
			logrus.Info("generated tts")
		}(alt)
	})

	go func() {
		after := time.After(time.Until(token.ExpiresAt))
		for {
			select {
			case <-ctx.Done():
				return
			case id := <-ch:
				if id == ctx.Config().TwitchBotID {
					data, err = r.Get(ctx, fmt.Sprintf("twitch:bot:%s", ctx.Config().TwitchBotID))
					if err != nil {
						logrus.WithError(err).Error("bad token from redis")
						continue
					}
					if err = json.UnmarshalFromString(data, &token); err != nil {
						logrus.WithError(err).Error("bad token from redis")
						continue
					}
				}
				continue
			case <-after:
			}
			token, err = RefreshToken(ctx, token.RefreshToken)
			if err != nil {
				logrus.WithError(err).Error("failed to refresh token")
				token.ExpiresAt = time.Now().Add(time.Minute * 10)
				continue
			}
			client.cl.SetIRCToken(fmt.Sprintf("oauth:%s", token.AccessToken))
			after = time.After(time.Until(token.ExpiresAt))
		}
	}()

	go func() {
		if err = client.cl.Connect(); err != nil {
			logrus.WithError(err).Fatal("twitch failed")
		}
	}()

	logrus.Infoln("connected and running.")

	return client, nil
}

func (c *twitchClient) SendMessage(channel string, message string) error {
	c.cl.Say(channel, message)
	return nil
}

func (c *twitchClient) SendWhisper(username string, message string) error {
	c.cl.Whisper(username, message)
	return nil
}

func RefreshToken(ctx global.Context, refreshToken string) (TokenResp, error) {
	body, err := json.Marshal(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     ctx.Config().TwitchClientID,
		"client_secret": ctx.Config().TwitchClientSecret,
	})
	if err != nil {
		return TokenResp{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://id.twitch.tv/oauth2/token", bytes.NewBuffer(body))
	if err != nil {
		return TokenResp{}, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TokenResp{}, err
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return TokenResp{}, err
	}

	token := TokenResp{}
	err = json.Unmarshal(body, &token)
	return token, err
}
