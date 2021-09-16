package global

import (
	"github.com/troydota/tts-textparser/src/instances"
)

type ServerCfg struct {
	ConfigFile string `mapstructure:"config_file"`
	Level      string `mapstructure:"level"`

	TtsChannelID string `mapstructure:"tts_channel_id"`

	RedisURI         string `mapstructure:"redis_uri"`
	RedisTaskSetKey  string `mapstructure:"redis_task_set_key"`
	RedisOutputEvent string `mapstructure:"redis_output_event"`

	MongoURI string `mapstructure:"mongo_uri"`
	MongoDB  string `mapstructure:"mongo_db"`

	StreamElementsEnabled    bool   `mapstructure:"stream_elements_enabled"`
	StreamElementsWssUrl     string `mapstructure:"stream_elements_wss_url"`
	StreamElementsAuthToken  string `mapstructure:"stream_elements_auth_token"`
	StreamElementsAuthMethod string `mapstructure:"stream_elements_auth_method"`

	ApiBind string `mapstructure:"api_bind"`

	Cors []string `mapstructure:"cors"`

	TwitchClientID     string `mapstructure:"twitch_client_id"`
	TwitchClientSecret string `mapstructure:"twitch_client_secret"`
	TwitchRedirectURI  string `mapstructure:"twitch_redirect_uri"`

	CookieDomain string `mapstructure:"cookie_domain"`
	CookieSecure bool   `mapstructure:"cookie_secure"`

	JwtSecret string `mapstructure:"jwt_secret"`

	TwitchBotID               string   `mapstructure:"twitch_bot_id"`
	TwitchBotUsername         string   `mapstructure:"twitch_bot_username"`
	TwitchBotControlChannel   string   `mapstructure:"twitch_bot_control_channel"`
	WhitelistedTwitchAccounts []string `mapstructure:"whitelisted_twitch_accounts"`

	FrontendDomain string `mapstructure:"frontend_domain"`

	mongo instances.MongoInstance
	redis instances.RedisInstance
	tts   instances.TtsInstance
}
