package processing

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"slices"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"

	cfg "tf-plan-reporter/internal/config"

	tfJson "github.com/hashicorp/terraform-json"
)

var (
	reportTable *consolidatedJson = new(consolidatedJson)

	//go:embed templates
	content embed.FS
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

			for _, resource := range tfPlan.ResourceChanges {
				resourceItem := &resourceData{
					resourceType:  resource.Type,
					resourceName:  resource.Name,
					resourceIndex: resource.Index,
				}

				tableRecordContext := log.WithFields(log.Fields{
					"resource_type":  resourceItem.resourceType,
					"resource_name":  resourceItem.resourceName,
					"resource_index": resourceItem.resourceIndex,
				})
				tableRecordContext.Debug("Created new resource item of report table")

				switch {
				case slices.Contains(resource.Change.Actions, tfJson.ActionCreate):
					reportTable.created = append(reportTable.created, resourceItem)

					tableRecordContext.Debug("The item has been put to 'Created' list")

				case slices.Contains(resource.Change.Actions, tfJson.ActionDelete):
					reportTable.deleted = append(reportTable.deleted, resourceItem)

					tableRecordContext.Debug("The item has been put to 'Deleted' list")

				case slices.Contains(resource.Change.Actions, tfJson.ActionUpdate):
					reportTable.updated = append(reportTable.updated, resourceItem)

					tableRecordContext.Debug("The item has been put to 'Updated' list")

				case slices.Contains(resource.Change.Actions, tfJson.ActionNoop):
					reportTable.unchanged = append(reportTable.unchanged, resourceItem)

					tableRecordContext.Debug("The item has been put to 'Unchanged' list")
				}
			}

		}
	}
}

// PrintReport function prepares and print the report from the data collected by function RunSearch
func PrintReport() {
	filename := cfg.AppConfig.ReportFileName
	var output io.Writer

	if len(filename) > 0 {
		var err error
		if output, err = os.Create(filename); err != nil {
			log.Fatal(err)
		}

		log.WithField("file_name", filename).Debug("The empty report file has been created")

		defer (output.(io.Closer)).Close()
	} else {
		output = os.Stdout

		log.Debug("The report is going to be printed to Stdout")
	}

	log.WithField("total_amount", reportTable.totalItems()).Debug("Report table contains elements")

	if !reportTable.isEmpty() {
		funcMap := template.FuncMap{
			"upper": strings.ToUpper,
		}
		tmpl := template.Must(template.New("default.tmpl").Funcs(funcMap).ParseFS(content, "templates/default.tmpl"))
		type reportData struct {
			WarningMarker string
			Caption       string
			ItemCount     int
			MainContent   string
		}

		processQueue := map[string][]*resourceData{
			"delete": reportTable.deleted,
			"create": reportTable.created,
			"update": reportTable.updated,
			"noop":   reportTable.unchanged,
		}

		headers := map[string][]string{
			"delete": []string{":red_circle:", "Resources to be deleted"},
			"create": []string{":green_circle:", "Resources to be created"},
			"update": []string{":orange_circle:", "Resources to be updated"},
			"noop":   []string{":gray_circle:", "Resources to be ignored for change"},
		}

		for key, value := range processQueue {

			tableContext := log.WithFields(log.Fields{
				"action_type":        key,
				"affected_resources": len(value),
			})
			tableContext.Debug("Statistics")

			if len(value) > 0 {

				tableContext.Debug("Preparing of resource table")

				tmplData := &reportData{
					ItemCount:     len(value),
					MainContent:   formatMainContent(value, key == "delete").String(),
					WarningMarker: headers[key][0],
					Caption:       headers[key][1],
				}

				tableContext.Debug("Output of resource table")
				if err := tmpl.Execute(output, tmplData); err != nil {
					tableContext.Fatal(err)
				}
			}
		}
	} else {
		fmt.Fprint(output, "THERE IS NO ANY REPORT DATA")
	}
}
