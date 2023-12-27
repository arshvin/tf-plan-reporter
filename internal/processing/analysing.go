package processing

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"

	app "tf-plan-reporter/internal/config"

	tfJson "github.com/hashicorp/terraform-json"
)

var (
	reportTable *consolidatedJson = new(consolidatedJson)

	//go:embed templates
	content embed.FS
	markers = map[string]string{
		"deleted":   ":red_circle:",
		"created":   ":green_circle:",
		"updated":   ":orange_circle:",
		"unchanged": ":gray_circle:",
	}
)

type processingRequest struct {
	commandName string
	planPath    string
	parsedData  chan<- tfJson.Plan
	pool        chan int
}

func findAllTFPlanFiles(rootDir string, fileBasename string) []string {
	var result []string
	if !path.IsAbs(rootDir) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Could not get current working dir")
		}
		rootDir = path.Join(cwd, rootDir)
	}

	if err := filepath.WalkDir(rootDir, func(currentPath string, d fs.DirEntry, err error) error {

		if d.Type().IsRegular() {
			pathElements := strings.Split(currentPath, string(os.PathSeparator))

			if pathElements[len(pathElements)-1] == fileBasename {
				log.Debugf("Found TF plan file: %s", currentPath)

				result = append(result, currentPath)
			}
		}

		return nil
	}); err != nil {
		log.Panicf("During directory tree walking the error happened: %s", err)
	}

	return result
}

// RunSearch function does:
// 1. searches all terraform generated binary plan files, with basename specified in `terraform_plan_file_basename`,
// starting from root director specified in `terraform_plan_search_folder` config file parameter
// 2. fills of `reportTable` variable, by parsed terraform plan data, which further is going to be source of printed report
func RunSearch(config app.ConfigFile) {
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

		//Save current working dir for return it back
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Could not get current working dir")
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

		//Return saved earlier current working dir
		if err = os.Chdir(cwd); err != nil {
			log.Fatal("Could not return back original working directory")
		}
	}
}

// PrintReport function prepares and print the report from the data collected by function RunSearch
func PrintReport(filename string) {
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
		tmpl := template.Must(template.New("default.tmpl").ParseFS(content, "templates/default.tmpl"))
		type reportData struct {
			WarningMarker string
			ItemCount     int
			MainContent   string
		}

		processQueue := map[string][]*resourceData{
			"deleted":   reportTable.deleted,
			"created":   reportTable.created,
			"updated":   reportTable.updated,
			"unchanged": reportTable.unchanged,
		}

		for key, value := range processQueue {

			tableContext := log.WithFields(log.Fields{
				"resource_type": key,
				"total_amount":  len(value),
			})
			tableContext.Debug("Statistics")

			if len(value) > 0 {

				tableContext.Debug("Preparing of resource table")

				tmplData := &reportData{
					ItemCount:     len(value),
					MainContent:   formatMainContent(value).String(),
					WarningMarker: markers[key],
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

func tfPlanReader(pr *processingRequest) {
	planFileContext := log.WithField("plan_file_name", pr.planPath)
	planFileContext.Printf("Preparation for parsing")

	planFileContext.Debug("Waiting of green light in process pool")
	pr.pool <- 1
	planFileContext.Debug("Green light has been acquired")

	if err := os.Chdir(path.Dir(pr.planPath)); err != nil {
		planFileContext.Panicf("Could not change current working dir: %s", err)
	}

	cmdResolvedPath, err := exec.LookPath(pr.commandName)
	if err != nil {
		planFileContext.Fatalf("Could not find the command file: %s", pr.commandName)
	}

	cmdContext := planFileContext.WithFields(log.Fields{
		"cwd":            path.Dir(pr.planPath),
		"command":        cmdResolvedPath,
		"plan_file_name": path.Base(pr.planPath),
	})
	cmdContext.Debug("Launching of command")

	cmd := exec.Command(cmdResolvedPath, "show", "-json", "-no-color", path.Base(pr.planPath))
	var outputPlan strings.Builder
	cmd.Stdout = &outputPlan

	err = cmd.Run()
	if err != nil {
		cmdContext.Fatalf("During execution the error happened: %s", err)
	}

	var tfJsonPlan tfJson.Plan
	err = tfJsonPlan.UnmarshalJSON([]byte(outputPlan.String()))
	if err != nil {
		planFileContext.Fatalf("Could not unmarshal: %s", err)
	}

	pr.parsedData <- tfJsonPlan

	planFileContext.Print("Parsing finished")
	//Return back capacity to the pool
	<-pr.pool
}
