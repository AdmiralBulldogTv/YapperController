package twitch

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/admiralbulldogtv/yappercontroller/src/textparser"
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

	client.cl.Join(ctx.Config().TwitchBotControlChannel)
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
			_, err := ctx.GetTtsInstance().Generate(ctx, msg, primitive.NewObjectIDFromTimestamp(time.Now()), channelID, textparser.Voices[0], textparser.Voices)
			if err != nil {
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
		}
	})

	client.cl.OnPrivateMessage(func(message twitch.PrivateMessage) {
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
			_, err := ctx.GetTtsInstance().Generate(ctx, msg, primitive.NewObjectIDFromTimestamp(time.Now()), channelID, textparser.Voices[0], textparser.Voices)
			if err != nil {
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
		}
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
