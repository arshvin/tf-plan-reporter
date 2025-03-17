package config

import (
	"slices"
	"strings"
)

type ConfigFile struct {
	BinaryFile        string   `mapstructure:"terraform_binary_file"`
	PlanFileBasename  string   `mapstructure:"terraform_plan_file_basename"`
	SearchFolder      string   `mapstructure:"terraform_plan_search_folder"`
	CriticalResources []string `mapstructure:"critical_resources"`
	AllowedRemovals   []string `mapstructure:"allowed_removals"`
}

type DefensePlan struct {
	IsAllCriticalSpecified bool
	RescueList             map[string]bool
	IgnoreList             map[string]bool
}

type MergedConfig struct {
	ConfigFile
	ReportFileName         string
	FailIfCriticalRemovals bool
	CriticalRemovalsFound  bool
	DefensePlan
}

var AppConfig *MergedConfig

func init() {
	AppConfig = new(MergedConfig)
	AppConfig.RescueList = make(map[string]bool)
	AppConfig.IgnoreList = make(map[string]bool)
}

func defineCriticalResources() {
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
