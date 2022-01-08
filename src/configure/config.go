package configure

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func checkErr(err error) {
	if err != nil {
		logrus.WithError(err).Fatal("config")
	}
}

func New() *Config {
	config := viper.New()

	// Default config
	b, _ := json.Marshal(Config{
		ConfigFile: "config.yaml",
	})
	tmp := viper.New()
	defaultConfig := bytes.NewReader(b)
	tmp.SetConfigType("json")
	checkErr(tmp.ReadConfig(defaultConfig))
	checkErr(config.MergeConfigMap(viper.AllSettings()))

	pflag.String("config", "config.yaml", "Config file location")
	pflag.Bool("noheader", false, "Disable the startup header")
	pflag.Parse()
	checkErr(config.BindPFlags(pflag.CommandLine))

	// File
	config.SetConfigFile(config.GetString("config"))
	config.AddConfigPath(".")
	err := config.ReadInConfig()
	if err != nil {
		logrus.Warning(err)
		logrus.Info("Using default config")
	} else {
		checkErr(config.MergeInConfig())
	}

	BindEnvs(config, Config{})

	// Environment
	config.AutomaticEnv()
	config.SetEnvPrefix("YAPPER")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AllowEmptyEnv(true)

	// Print final config
	c := &Config{}
	checkErr(config.Unmarshal(&c))

	initLogging(c.Level)

	return c
}

func BindEnvs(config *viper.Viper, iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		switch v.Kind() {
		case reflect.Struct:
			BindEnvs(config, v.Interface(), append(parts, tv)...)
		default:
			_ = config.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}

type Config struct {
	ConfigFile string `mapstructure:"config_file"`
	Level      string `mapstructure:"level"`

	TtsChannelID string `mapstructure:"tts_channel_id"`

	Redis struct {
		URI         string `mapstructure:"uri"`
		TaskSetKey  string `mapstructure:"task_set_key"`
		OutputEvent string `mapstructure:"output_event"`
	} `mapstructure:"redis"`

	Mongo struct {
		URI      string `mapstructure:"uri"`
		Database string `mapstructure:"database"`
	} `mapstructure:"mongo"`

	StreamElements struct {
		Enabled    bool   `mapstructure:"enabled"`
		WssURL     string `mapstructure:"wss_url"`
		AuthToken  string `mapstructure:"auth_token"`
		AuthMethod string `mapstructure:"auth_method"`
	} `mapstructure:"streamelements"`

	Twitch struct {
		ClientID            string   `mapstructure:"client_id"`
		ClientSecret        string   `mapstructure:"client_secret"`
		RedirectURI         string   `mapstructure:"redirect_uri"`
		BotID               string   `mapstructure:"bot_id"`
		BotUsername         string   `mapstructure:"bot_username"`
		BotControlChannel   string   `mapstructure:"bot_control_channel"`
		StreamerChannel     string   `mapstructure:"streamer_channel"`
		WhitelistedAccounts []string `mapstructure:"whitelisted_accounts"`
	} `mapstructure:"twitch"`

	CookieDomain string   `mapstructure:"cookie_domain"`
	CookieSecure bool     `mapstructure:"cookie_secure"`
	Cors         []string `mapstructure:"cors"`
	ApiBind      string   `mapstructure:"api_bind"`

	JwtSecret string `mapstructure:"jwt_secret"`

	FrontendDomain string `mapstructure:"frontend_domain"`
}
