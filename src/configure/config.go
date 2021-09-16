package configure

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/troydota/tts-textparser/src/global"
)

// default config
var defaultConf = global.ServerCfg{
	ConfigFile: "config.yaml",
}

var Config = viper.New()

func initLog() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	if l, err := log.ParseLevel(Config.GetString("level")); err == nil {
		log.SetLevel(l)
	}
}

func checkErr(err error) {
	if err != nil {
		log.WithError(err).Fatal("failed on configure")
	}
}

func Init(ctx context.Context) global.Context {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)

	// Default config
	b, _ := json.Marshal(defaultConf)
	defaultConfig := bytes.NewReader(b)
	viper.SetConfigType("json")
	checkErr(viper.ReadConfig(defaultConfig))
	checkErr(Config.MergeConfigMap(viper.AllSettings()))

	// Environment
	replacer := strings.NewReplacer(".", "_")
	Config.SetEnvKeyReplacer(replacer)
	Config.AllowEmptyEnv(true)
	Config.AutomaticEnv()

	// File
	Config.SetConfigFile(Config.GetString("config_file"))
	Config.AddConfigPath(".")
	err := Config.ReadInConfig()
	if err != nil {
		log.Warning(err)
		log.Info("Using default config")
	} else {
		checkErr(Config.MergeInConfig())
	}

	// Log
	initLog()

	// Print final config
	c := global.ServerCfg{}
	checkErr(Config.Unmarshal(&c))

	return global.NewCtx(ctx, c)
}
