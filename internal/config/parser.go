package config

import (
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// TODO: implement test that config file is parsed with different values correctly: optional, mandatory, etc.
func Parse(name string) *AppConfig {
	viper_runtime := viper.New()

	viper_runtime.SetConfigType("yaml")
	viper_runtime.SetConfigFile(name)

	if err := viper_runtime.ReadInConfig(); err != nil {
		log.Panic(err)
	}

	appConfig := create()
	var configFile ConfigFile
	if err := viper_runtime.Unmarshal(&configFile); err != nil {
		log.Panic(err)
	}

	log.Debugf("Content of config file/structure: %v", configFile)
	appConfig.ConfigFile = configFile

	if i := slices.IndexFunc(appConfig.ConfigFile.CriticalResources, func(s string) bool {
		return strings.TrimSpace(strings.ToLower(s)) == "all"
	}); i > -1 {
		appConfig.IsAllCriticalSpecified = true
		for _, item := range appConfig.AllowedRemovals {
			appConfig.IgnoreList[strings.ToLower(item)] = true
		}
	} else {
		appConfig.IsAllCriticalSpecified = false
		for _, item := range appConfig.CriticalResources {
			appConfig.RescueList[strings.ToLower(item)] = true
		}
	}

	return appConfig
}
