package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func ProcessFileConfig(name string) {
	viper_runtime := viper.New()

	viper_runtime.SetConfigType("yaml")
	viper_runtime.SetConfigFile(name)

	if err := viper_runtime.ReadInConfig(); err != nil {
		log.Panic(err)
	}

	var config ConfigFile
	if err := viper_runtime.Unmarshal(&config); err != nil {
		log.Panic(err)
	}

	log.Debugf("Content of config file/structure: %v", config)

	AppConfig.ConfigFile = config

	defineCriticalResources()
}
