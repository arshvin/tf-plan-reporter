package processing

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	log "github.com/sirupsen/logrus"

	cmn "github.com/arshvin/tf-plan-reporter/internal"
	cfg "github.com/arshvin/tf-plan-reporter/internal/config"
	"github.com/arshvin/tf-plan-reporter/internal/report"

	tfJson "github.com/hashicorp/terraform-json"
)

var (
	reportTable *cmn.ConsolidatedJson = new(cmn.ConsolidatedJson)
)

type processingRequest struct {
	commandName string
	planPath    string
	parsedData  chan<- tfJson.Plan
	pool        chan int
}

// RunSearch function does:
// 1. searches all terraform generated binary plan files, with basename specified in `terraform_plan_file_basename`,
// starting from root director specified in `terraform_plan_search_folder` config file parameter
// 2. fills of `reportTable` variable, by parsed terraform plan data, which further is going to be source of printed report
func RunSearch() {
	config := cfg.AppConfig
	tfPlanFilesPathList := findAllTFPlanFiles(config.SearchFolder, config.PlanFileBasename)

	foundItems := len(tfPlanFilesPathList)
	log.WithFields(log.Fields{
		"plan_basename": config.PlanFileBasename,
		"total_amount":  foundItems,
	}).Debug("Found terraform generated plan files")

	if foundItems > 0 {

		pool := make(chan int, runtime.GOMAXPROCS(0))
		dataPipe := make(chan tfJson.Plan, runtime.GOMAXPROCS(0))

		absCmdBinaryPath := config.BinaryFile
		if !path.IsAbs(config.BinaryFile) {
			cwd, _ := os.Getwd()
			absCmdBinaryPath = path.Join(cwd, config.BinaryFile)
		}

		for _, absTFPlanFilePath := range tfPlanFilesPathList {
			pr := &processingRequest{
				commandName: absCmdBinaryPath,
				planPath:    absTFPlanFilePath,
				parsedData:  dataPipe,
				pool:        pool,
			}

			go tfPlanReader(pr)
		}

		//Parsing of read data
		log.Debug("Waiting of data from read TF plan files for processing")
		for item := 0; item < foundItems; item++ {
			tfPlan := <-dataPipe

			reportTable.Parse(&tfPlan)
		}
	}
}

// PrintReport function prepares and print the report from the data collected by function RunSearch
func PrintReport() {
	filename := cfg.AppConfig.ReportFileName
	var output io.Writer
	var reports []*report.Report

	totalAmount := reportTable.TotalItems()
	log.WithField("total_amount", totalAmount).Debug("Report table contains elements")

	if totalAmount > 0 {

		if len(filename) > 0 {
			var err error
			if output, err = os.Create(filename); err != nil {
				log.Fatal(err)
			}

			log.WithField("file_name", filename).Debug("The empty report file has been created")

			defer (output.(io.Closer)).Close()

			reports = []*report.Report{
				report.ForGitHub(output),
				report.ForStdout(),
			}

		} else {
			reports = []*report.Report{
				report.ForStdout(),
			}

			log.Debug("The report is going to be printed to Stdout")
		}

		for _, r := range reports {
			r.Prepare(reportTable)
			r.Print()
		}

		if cfg.AppConfig.FailIfCriticalRemovals && cfg.AppConfig.CriticalRemovalsFound {
			log.Fatal("There are critical resources removal in the report")
		}

	} else {
		fmt.Fprint(output, "THERE IS NO ANY REPORT DATA")
	}
}
