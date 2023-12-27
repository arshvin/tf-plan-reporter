package internal

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ConfigFile struct {
	BinaryFile       string `mapstructure:"terraform_binary_file"`
	PlanFileBasename string `mapstructure:"terraform_plan_file_basename"`
	SearchFolder     string `mapstructure:"terraform_plan_search_folder"`
}

func ProcessFileConfig(name string) ConfigFile {
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

	return config
}
