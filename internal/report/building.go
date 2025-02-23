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
	var output io.Writer
	var reports []*report

	totalAmount := reportData.TotalItems()
	log.WithField("total_amount", totalAmount).Debug("Report table contains elements")

	if totalAmount > 0 {

		if len(outputFilename) > 0 {
			var err error
			if output, err = os.Create(outputFilename); err != nil {
				log.Fatal(err)
			}

			log.WithField("file_name", outputFilename).Debug("The empty report file has been created")

			defer (output.(io.Closer)).Close()

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
		fmt.Fprint(output, "THERE IS NO ANY REPORT DATA")
	}
}
