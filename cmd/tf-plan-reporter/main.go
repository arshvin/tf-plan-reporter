package main

import (
	"flag"
	app "tf-plan-reporter/internal/config"
	analysis "tf-plan-reporter/internal/processing"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	config_filename_flag string = "config-file"
	print_config_flag    string = "print"
	report_file_name string =  "report-file"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		PadLevelText: true,
		DisableTimestamp: false,
		FullTimestamp: true,
		ForceColors: true,
	})
}

func main() {

	flag.String(config_filename_flag, "", "Config file name of the App")
	flag.Bool(print_config_flag, false, "Print the example of the config file")
	flag.String(report_file_name, "", "Output report file name")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if viper.IsSet(config_filename_flag) {
		config_filename := viper.GetString(config_filename_flag)
		appConfig := app.ProcessFileConfig(config_filename)
		analysis.RunSearch(appConfig)
		analysis.PrintReport(viper.GetString((report_file_name)))
	}

	if viper.IsSet(print_config_flag) {
		print_config := viper.GetString(print_config_flag)
		log.Printf("Output to screen the example of config: %s", print_config)
	}
}
