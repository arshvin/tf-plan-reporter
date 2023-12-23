package main

import (
	"flag"
	"log"
	app "tf-plan-reporter/internal/config"
	analysis "tf-plan-reporter/internal/processing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	config_filename_flag string = "config-file"
	print_config_flag    string = "print"
	report_file_name string =  "report-file"
)

var (
	logger = log.Default()
)

func init() {
	logger.SetPrefix("INFO: ")
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
		logger.Printf("Output to screen the example of config: %s", print_config)
	}
}
