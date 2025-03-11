package config

type ConfigFile struct {
	TfCmdBinaryFile    string   `mapstructure:"terraform_binary_file"`
	TfPlanFileBasename string   `mapstructure:"terraform_plan_file_basename"`
	SearchFolder       string   `mapstructure:"terraform_plan_search_folder"`
	CriticalResources  []string `mapstructure:"critical_resources"`
	AllowedRemovals    []string `mapstructure:"allowed_removals"`
	NotUseTfChDirArg   bool     `mapstructure:"not_use_chdir,omitempty"`
}

type DefensePlan struct {
	IsAllCriticalSpecified bool
	RescueList             map[string]bool
	IgnoreList             map[string]bool
}

type AppConfig struct {
	ConfigFile
	ReportFileName         string
	FailIfCriticalRemovals bool
	DefensePlan
}

func create() *AppConfig {
	appCfg := new(AppConfig)
	appCfg.RescueList = make(map[string]bool)
	appCfg.IgnoreList = make(map[string]bool)
	return appCfg
}
