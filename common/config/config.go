package config

import (
	"github.com/fsnotify/fsnotify"
	_ "github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
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
	conf.AddConfigPath("./set")
	err := conf.ReadInConfig()
	if err != nil {
		logrus.WithField("config", "conf").WithError(err).Panicf("unable to read global config")
	}
	conf.WatchConfig()
	conf.OnConfigChange(func(e fsnotify.Event) {
		err := conf.ReadInConfig()
		if err != nil {
			logrus.WithField("config", "conf").Info("config file update; change: ", e.Name)
		}
	})
}
