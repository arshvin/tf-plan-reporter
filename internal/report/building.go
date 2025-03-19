package report

import (
	"fmt"
	"io"
	"os"

	"github.com/arshvin/tf-plan-reporter/internal/processing"
	log "github.com/sirupsen/logrus"
)

// PrintReport function prepares and print the report from the data collected by function RunSearch
func PrintReport(reportData *processing.ConsolidatedJson, outputFilename string) {

	totalAmount := reportData.TotalItems()
	log.WithField("total_amount", totalAmount).Debug("Report table contains elements")

	if totalAmount > 0 {
		var reports []*report

		if len(outputFilename) > 0 {
			output, err := os.Create(outputFilename)

			if err != nil {
				log.Fatal(err)
			}

			log.WithField("file_name", outputFilename).Debug("The empty report file has been created")

			defer output.Close()

			reports = []*report{
				forGitHub(output),
				forStdout(),
			}

		} else {
			reports = []*report{
				forStdout(),
			}

			log.Debug("The report is going to be printed to Stdout")
		}

		for _, r := range reports {
			r.Prepare(reportData)
			r.Print()
		}

	} else {
		fmt.Print("THERE IS NO ANY REPORT DATA")
	}
}
