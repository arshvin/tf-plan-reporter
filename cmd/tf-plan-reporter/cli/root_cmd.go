package cli

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/arshvin/tf-plan-reporter/internal/config"
	"github.com/arshvin/tf-plan-reporter/internal/processing"
	"github.com/arshvin/tf-plan-reporter/internal/report"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	configFileArg         = "config-file"
	printConfigExampleArg = "print-example"
)

var (
	configFileName         string
	outputFileName         string
	onlyPrintConfigExample bool
	exitWithError          bool
	debugOutput            bool
	noColor                bool
)

// Entry point of tf-plan-reporter tool
func Execute() {

	flag.StringVar(&configFileName, configFileArg, "", "Config file name of the App")
	flag.StringVar(&outputFileName, "report-file", "", "Output file name of the report ")
	flag.BoolVar(&onlyPrintConfigExample, printConfigExampleArg, false, "Print an example of the App config file without analyses run")
	flag.BoolVar(&exitWithError, "keep-gate", false, "Finish App with non zero exit code if critical resources removals are detected")

	flag.BoolVar(&debugOutput, "verbose", false, "Add debug logging output")
	flag.BoolVar(&noColor, "no-color", false, "Turn off color output in log messages")

	flag.Bool("help", false, "help message output")
	flag.Bool("h", false, "help message output")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine.MarkHidden("help")
	pflag.CommandLine.MarkHidden("h")

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

	if ok, _ := pflag.CommandLine.GetBool("help"); ok {
		pflag.Usage()
		os.Exit(0) //Explicitly
	}

	if onlyPrintConfigExample {
		config.PrintExample()
		os.Exit(0) //Explicitly
	}

	if len(configFileName) > 0 {
		settings := config.Parse(configFileName)

		//Similar checking, if settings.NotUseTfChDirArg == false, will be further once all tf-plan files found
		if settings.NotUseTfChDirArg {
			log.Debug("Checking if Terraform providers folder exists in current folder in advance")

			cwd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			if !processing.TfProviderFolderExist(cwd){
				log.Fatal("Terraform providers folder was not found in current working directory")
			}
		}

		settings.ReportFileName = outputFileName
		settings.FailIfCriticalRemovals = exitWithError

		collectedData := processing.CollectBinaryData(
			settings.SearchFolder,
			settings.TfPlanFileBasename,
			settings.TfCmdBinaryFile,
			settings.NotUseTfChDirArg,
		)

		dm:=processing.GetDecisionMaker()
		dm.SetConfig(settings)

		report.PrintReport(collectedData,settings.ReportFileName)

		if settings.FailIfCriticalRemovals && dm.CriticalRemovalsFound() {
			log.Fatal("There are critical resources removal in the report")
		}

		os.Exit(0) //Explicitly
	}

	pflag.Usage()
	log.Fatalf("At least one of the following flags must be chosen: %s, %s", configFileArg, printConfigExampleArg)
}
