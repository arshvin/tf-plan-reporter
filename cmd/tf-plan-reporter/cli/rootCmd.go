package cli

import (
	"flag"
	"log"

	app "tf-plan-reporter/internal/config"
	analysis "tf-plan-reporter/internal/processing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	configFilenameFlag      string = "config-file"
	printConfigFlag         string = "print"
	reportFileNameFlag      string = "report-file"
	errorIfCriticalRemovals string = "fail-if"
)

func Execute() {
	flag.String(configFilenameFlag, "", "Config file name of the App")
	flag.Bool(printConfigFlag, false, "Print the example of the config file")
	flag.String(reportFileNameFlag, "", "Output report file name")
	flag.Bool(errorIfCriticalRemovals, false, "Return error if there are resources deleting in 'critical_resources' or not specified in 'allowed_removals config lists'")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if viper.IsSet(configFilenameFlag) {
		config_filename := viper.GetString(configFilenameFlag)
		app.ProcessFileConfig(config_filename)

		app.AppConfig.ReportFileName = viper.GetString(reportFileNameFlag)
		app.AppConfig.FailIfCriticalRemovals = viper.GetBool(errorIfCriticalRemovals)

		analysis.RunSearch()
		analysis.PrintReport()
	}

	if viper.IsSet(printConfigFlag) {
		print_config := viper.GetString(printConfigFlag)
		log.Printf("Output to screen the example of config: %s", print_config)
	}
}
