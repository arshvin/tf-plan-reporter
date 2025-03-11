package config

import (
	"os"
	"path"
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
	//FIXME: Figure out what happens here if configFile is not valid
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

	//Replacing of relative TF command path to absolute one if it's required
	if !path.IsAbs(appConfig.TfCmdBinaryFile) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Could not get current working dir")
		}
		appConfig.TfCmdBinaryFile = path.Join(cwd, appConfig.TfCmdBinaryFile)
	}

	//Replacing of relative SearchFolder path to absolute one if it's required
	if !path.IsAbs(appConfig.SearchFolder) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Could not get current working dir")
		}
		appConfig.SearchFolder = path.Join(cwd, appConfig.SearchFolder)
	}

	return appConfig
}
