package config

import (
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Parse(name string) {
	viper_runtime := viper.New()

	viper_runtime.SetConfigType("yaml")
	viper_runtime.SetConfigFile(name)

	if err := viper_runtime.ReadInConfig(); err != nil {
		log.Panic(err)
	}

	//FIXME: Figure out what happens here if config is not valid
	var config ConfigFile
	if err := viper_runtime.Unmarshal(&config); err != nil {
		log.Panic(err)
	}

	log.Debugf("Content of config file/structure: %v", config)

	AppConfig.ConfigFile = config

	if i := slices.IndexFunc(AppConfig.ConfigFile.CriticalResources, func(s string) bool {
		return strings.TrimSpace(strings.ToLower(s)) == "all"
	}); i > -1 {
		AppConfig.IsAllCriticalSpecified = true
		for _, item := range AppConfig.AllowedRemovals {
			AppConfig.IgnoreList[strings.ToLower(item)] = true
		}
	} else {
		AppConfig.IsAllCriticalSpecified = false
		for _, item := range AppConfig.CriticalResources {
			AppConfig.RescueList[strings.ToLower(item)] = true
		}
	}
}
