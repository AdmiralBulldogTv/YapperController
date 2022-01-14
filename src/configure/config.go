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
	ConfigFile string `mapstructure:"config_file" json:"config_file"`
	Level      string `mapstructure:"level" json:"level"`

	TtsChannelID string `mapstructure:"tts_channel_id" json:"tts_channel_id"`

	Redis struct {
		Username    string   `mapstructure:"username" json:"username"`
		Password    string   `mapstructure:"password" json:"password"`
		MasterName  string   `mapstructure:"master_name" json:"master_name"`
		Addresses   []string `mapstructure:"addresses" json:"addresses"`
		Database    int      `mapstructure:"database" json:"database"`
		Sentinel    bool     `mapstructure:"sentinel" json:"sentinel"`
		TaskSetKey  string   `mapstructure:"task_set_key" json:"task_set_key"`
		OutputEvent string   `mapstructure:"output_event" json:"output_event"`
	} `mapstructure:"redis" json:"redis"`

	Mongo struct {
		URI      string `mapstructure:"uri" json:"uri"`
		Database string `mapstructure:"database" json:"database"`
	} `mapstructure:"mongo" json:"mongo"`

	StreamElements struct {
		Enabled    bool   `mapstructure:"enabled" json:"enabled"`
		WssURL     string `mapstructure:"wss_url" json:"wss_url"`
		AuthToken  string `mapstructure:"auth_token" json:"auth_token"`
		AuthMethod string `mapstructure:"auth_method" json:"auth_method"`
	} `mapstructure:"streamelements" json:"streamelements"`

	Twitch struct {
		ClientID            string   `mapstructure:"client_id" json:"client_id"`
		ClientSecret        string   `mapstructure:"client_secret" json:"client_secret"`
		RedirectURI         string   `mapstructure:"redirect_uri" json:"redirect_uri"`
		BotID               string   `mapstructure:"bot_id" json:"bot_id"`
		BotUsername         string   `mapstructure:"bot_username" json:"bot_username"`
		BotControlChannel   string   `mapstructure:"bot_control_channel" json:"bot_control_channel"`
		StreamerChannel     string   `mapstructure:"streamer_channel" json:"streamer_channel"`
		WhitelistedAccounts []string `mapstructure:"whitelisted_accounts" json:"whitelisted_accounts"`
	} `mapstructure:"twitch" json:"twitch"`

	CookieDomain string   `mapstructure:"cookie_domain" json:"cookie_domain"`
	CookieSecure bool     `mapstructure:"cookie_secure" json:"cookie_secure"`
	Cors         []string `mapstructure:"cors" json:"cors"`
	ApiBind      string   `mapstructure:"api_bind" json:"api_bind"`

	JwtSecret string `mapstructure:"jwt_secret" json:"jwt_secret"`

	FrontendDomain string `mapstructure:"frontend_domain" json:"frontend_domain"`
}
