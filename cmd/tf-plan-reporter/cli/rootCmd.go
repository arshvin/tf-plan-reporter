package cli

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	cfg "github.com/arshvin/tf-plan-reporter/internal/config"
	analysis "github.com/arshvin/tf-plan-reporter/internal/processing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	configFileArg         = "config-file"
	printConfigExampleArg = "print-example"
)

func Execute() {
	var configFileName string
	var outputFileName string
	var onlyPrintConfigExample bool
	var exitWithError bool
	var debugOutput bool
	var noColor bool

	flag.StringVar(&configFileName, configFileArg, "", "Config file name of the App")
	flag.StringVar(&outputFileName, "report-file", "", "Output file name of the report ")
	flag.BoolVar(&onlyPrintConfigExample, printConfigExampleArg, false, "Print an example of the App config file without analyses run")
	flag.BoolVar(&exitWithError, "keep-gate", false, "Finish App with non zero exit code if critical resources removals are detected")

	flag.BoolVar(&debugOutput, "verbose", false, "Add debug logging output")
	flag.BoolVar(&noColor, "no-color", false, "Turn off color output in log messages")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if debugOutput {
		log.SetLevel(log.DebugLevel)
	}

	if noColor {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: true,
		})
	}

	if len(configFileName) > 0 {
		cfg.ProcessFileConfig(configFileName)

		cfg.AppConfig.ReportFileName = outputFileName
		cfg.AppConfig.FailIfCriticalRemovals = exitWithError

		analysis.RunSearch()
		analysis.PrintReport()

		os.Exit(0) //Explicitly
	}

	if onlyPrintConfigExample {
		cfg.PrintExample()
		os.Exit(0) //Explicitly
	}

	pflag.Usage()
	log.Fatalf("It must be chosen one of the following flags: %s, %s", configFileArg, printConfigExampleArg)
}
