package internal

import (
	"github.com/spf13/viper"
	"log"
)

var (
	logger = log.Default()
)

type ConfigFile struct {
	BinaryFile       string `mapstructure:"terraform_binary_file"`
	PlanFileBasename string `mapstructure:"terraform_plan_file_basename"`
	SearchFolder     string `mapstructure:"terraform_plan_search_folder"`
}

var config ConfigFile

func ProcessFileConfig(name string) ConfigFile  {
	viper_runtime := viper.New()

	viper_runtime.SetConfigType("yaml")
	viper_runtime.SetConfigFile(name)

	if err := viper_runtime.ReadInConfig(); err != nil {
		logger.Panic(err)
	}

	if err := viper_runtime.Unmarshal(&config); err != nil {
		logger.Panic(err)
	}
	return config
}
