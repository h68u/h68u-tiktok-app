package config

import (
	_ "github.com/fsnotify/fsnotify"
	_ "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type config struct {
	*viper.Viper
}

var conf *config

func init() {
	conf = &config{
		viper.New(),
	}
	conf.SetConfigName("app")
	conf.SetConfigType("yaml")
	conf.AddConfigPath(".")
	err := conf.ReadInConfig()
	if err != nil {
		panic(err)
	}
	conf.WatchConfig()
}
