package v1

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/admiralbulldogtv/yappercontroller/src/global"
	"github.com/admiralbulldogtv/yappercontroller/src/jwt"
	"github.com/admiralbulldogtv/yappercontroller/src/twitch"
	"github.com/admiralbulldogtv/yappercontroller/src/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type TwitchUserResp struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
}

type TwitchUserRespWrapper struct {
	Data []TwitchUserResp `json:"data"`
}

func Twitch(ctx global.Context, app fiber.Router) {
	login := app.Group("/login")

	loginFunc := func(c *fiber.Ctx, scopes []string) error {
		data, err := utils.GenerateRandomString(64)
		if err != nil {
			logrus.WithError(err).Error("failed to generate random bytes")
			return c.SendStatus(500)
		}

		token, err := jwt.Sign(ctx.Config().JwtSecret, data)
		if err != nil {
			logrus.WithError(err).Error("failed to generate jwt")
			return c.SendStatus(500)
		}

		q := url.Values{
			"client_id":     []string{ctx.Config().Twitch.ClientID},
			"redirect_uri":  []string{ctx.Config().Twitch.RedirectURI},
			"response_type": []string{"code"},
			"scope":         []string{strings.Join(scopes, " ")},
			"state":         []string{token},
		}

		c.Cookie(&fiber.Cookie{
			Name:     "twitch_csrf",
			Value:    data,
			HTTPOnly: true,
			Secure:   ctx.Config().CookieSecure,
			Domain:   ctx.Config().CookieDomain,
			Expires:  time.Now().Add(time.Minute * 10),
		})

		return c.Redirect(fmt.Sprintf("https://id.twitch.tv/oauth2/authorize?%s", q.Encode()))
	}

	login.Get("/", func(c *fiber.Ctx) error {
		return loginFunc(c, []string{})
	})

	login.Get("/bot", func(c *fiber.Ctx) error {
		return loginFunc(c, []string{"chat:edit", "chat:read", "whispers:read", "whispers:edit"})
	})

	app.Get("/callback", func(c *fiber.Ctx) error {
		data := c.Cookies("twitch_csrf")
		token := c.Query("state")

		outData := ""

		if err := jwt.Verify(ctx.Config().JwtSecret, token, &outData); err != nil {
			logrus.WithError(err).Error("failed to verify jwt")
			return c.SendStatus(400)
		}

		if data != outData {
			logrus.Error("jwt missmatch")
			return c.SendStatus(400)
		}

		code := c.Query("code")
		body, _ := json.Marshal(map[string]string{
			"client_id":     ctx.Config().Twitch.ClientID,
			"client_secret": ctx.Config().Twitch.ClientSecret,
			"code":          code,
			"grant_type":    "authorization_code",
			"redirect_uri":  ctx.Config().Twitch.RedirectURI,
		})
		req, err := http.NewRequestWithContext(c.Context(), http.MethodPost, "https://id.twitch.tv/oauth2/token", bytes.NewBuffer(body))
		if err != nil {
			logrus.WithError(err).Error("failed to fetch access token")
			return c.SendStatus(500)
		}

		req.Header.Add("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logrus.WithError(err).Error("failed to fetch access token")
			return c.SendStatus(500)
		}

		defer resp.Body.Close()

		tokenResp := twitch.TokenResp{}
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.WithError(err).Error("failed to fetch access token")
			return c.SendStatus(500)
		}

		if err = json.Unmarshal(body, &tokenResp); err != nil {
			logrus.WithError(err).Error("failed to fetch access token")
			return c.SendStatus(500)
		}

		tokenResp.ExpiresAt = time.Now().Add(time.Duration(float64(tokenResp.ExpiresIn)*0.7) * time.Second)

		req, err = http.NewRequestWithContext(c.Context(), http.MethodGet, "https://api.twitch.tv/helix/users", nil)
		if err != nil {
			logrus.WithError(err).Error("failed to fetch access token")
			return c.SendStatus(500)
		}

		req.Header.Add("Client-Id", ctx.Config().Twitch.ClientID)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenResp.AccessToken))

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			logrus.WithError(err).Error("failed to fetch access token")
			return c.SendStatus(500)
		}
		defer resp.Body.Close()

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.WithError(err).Error("failed to fetch access token")
			return c.SendStatus(500)
		}

		userResp := TwitchUserRespWrapper{}
		if err = json.Unmarshal(body, &userResp); err != nil {
			logrus.WithError(err).Error("failed to fetch user")
			return c.SendStatus(500)
		}

		if len(userResp.Data) != 1 {
			logrus.WithError(err).Error("failed to fetch user")
			return c.SendStatus(500)
		}

		user := userResp.Data[0]
		if user.ID == ctx.Config().Twitch.BotID {
			redis := ctx.Inst().Redis
			data, _ = json.MarshalToString(tokenResp)
			if err = redis.Set(c.Context(), fmt.Sprintf("twitch:bot:%s", user.ID), data, -1); err != nil {
				logrus.WithError(err).Error("failed to fetch set bot token")
				return c.SendStatus(500)
			}
			if err = redis.Publish(c.Context(), "events:twitch:bot:login", user.ID); err != nil {
				logrus.WithError(err).Error("failed to fetch set bot token")
				return c.SendStatus(500)
			}
		}

		token, err = jwt.Sign(ctx.Config().JwtSecret, user)
		if err != nil {
			logrus.WithError(err).Error("failed to create user token")
			return c.SendStatus(500)
		}

		c.Cookie(&fiber.Cookie{
			Name:    "tts_auth",
			Value:   token,
			Domain:  ctx.Config().CookieDomain,
			Secure:  ctx.Config().CookieSecure,
			Expires: time.Now().Add(time.Hour * 24 * 14),
		})

		return c.Redirect(ctx.Config().FrontendDomain)
	})
}
