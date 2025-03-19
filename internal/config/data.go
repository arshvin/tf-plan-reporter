package config

type ConfigFile struct {
	TfCmdBinaryFile    string   `mapstructure:"terraform_binary_file"`
	TfPlanFileBasename string   `mapstructure:"terraform_plan_file_basename"`
	SearchFolder       string   `mapstructure:"terraform_plan_search_folder"`
	CriticalResources  []string `mapstructure:"critical_resources"`
	AllowedRemovals    []string `mapstructure:"allowed_removals"`
	NotUseTfChDirArg   bool     `mapstructure:"not_use_chdir"`
}

type DefensePlan struct {
	IsAllCriticalSpecified bool
	ExceptionalResources             map[string]bool //Depending on value IsAllCriticalSpecified this list (actually map) is `allowed for removal` (if true), or `critical for keeping` (if false)
}

type AppConfig struct {
	ConfigFile
	ReportFileName         string
	FailIfCriticalRemovals bool
	DefensePlan
}

func create() *AppConfig {
	appCfg := new(AppConfig)
	appCfg.ExceptionalResources = make(map[string]bool)

	return appCfg
}
