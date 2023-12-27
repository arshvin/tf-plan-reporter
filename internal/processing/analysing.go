package processing

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"text/template"

	app "tf-plan-reporter/internal/config"

	tfJson "github.com/hashicorp/terraform-json"
)

var (
	logger                        = log.Default()
	reportTable *consolidatedJson = new(consolidatedJson)

	//go:embed templates
	content embed.FS
)

type processingRequest struct {
	commandName string
	planPath    string
	parsedData  chan<- tfJson.Plan
	pool        chan int
	done        bool
}

func findAllTFPlanFiles(rootDir string, fileBasename string) []string {
	var result []string
	if err := filepath.WalkDir(rootDir, func(currentPath string, d fs.DirEntry, err error) error {

		if d.Type().IsRegular() {
			pathElements := strings.Split(currentPath, string(os.PathSeparator))

			if pathElements[len(pathElements)-1] == fileBasename {
				result = append(result, currentPath)
			}
		}

		return nil
	}); err != nil {
		logger.Panicf("During walking directory tree the error happened: %s", err)
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
	logger.Printf("Found following amount of terraform generated plan files with name '%s': %d", config.PlanFileBasename, foundItems)

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
			logger.Fatal("Unable to get current working dir")
		}

		for index, absTFPlanFilePath := range tfPlanFilesPathList {
			pr := &processingRequest{
				commandName: absCmdBinaryPath,
				planPath:    absTFPlanFilePath,
				parsedData:  dataPipe,
				pool:        pool,
				done:        false,
			}

			if index == foundItems-1 {
				pr.done = true
			}

			go tfPlanReader(pr)
		}

		//Parsing of read data
		logger.Printf("Waiting the data for processing from read TF plans")
		for tfPlan := range dataPipe {
			resourceChanges := tfPlan.ResourceChanges

			for _, resource := range resourceChanges {
				resourceItem := &resourceData{
					resourceType:  resource.Type,
					resourceName:  resource.Name,
					resourceIndex: resource.Index,
				}

				logger.Printf("Created new resource item of report table: %s", resourceItem)

				switch {
				case slices.Contains(resource.Change.Actions, tfJson.ActionCreate):
					reportTable.created = append(reportTable.created, resourceItem)

					logger.Printf("It's going to be put to 'Created' list ")
				case slices.Contains(resource.Change.Actions, tfJson.ActionDelete):
					reportTable.deleted = append(reportTable.deleted, resourceItem)

					logger.Printf("It's going to be put to 'Deleted' list ")
				case slices.Contains(resource.Change.Actions, tfJson.ActionUpdate):
					reportTable.updated = append(reportTable.updated, resourceItem)

					logger.Printf("It's going to be put to 'Updated' list ")
				case slices.Contains(resource.Change.Actions, tfJson.ActionNoop):
					reportTable.unchanged = append(reportTable.unchanged, resourceItem)

					logger.Printf("It's going to be put to 'Unchanged' list ")
				}
			}
		}

		//Return saved earlier current working dir
		if err = os.Chdir(cwd); err != nil {
			logger.Fatal("Unable to return back original working directory")
		}
	}
}

// PrintReport function prepares and print the report from the data collected by function RunSearch
func PrintReport(filename string) {
	var output io.Writer

	if len(filename) > 0 {
		var err error
		if output, err = os.Create(filename); err != nil {
			logger.Fatal(err)
		}

		logger.Printf("The empty report files has been created: %s", filename)

		defer (output.(io.Closer)).Close()
	} else {
		output = os.Stdout
	}

	logger.Printf("Report table consists of amount of elements: %d", reportTable.totalItems())

	if !reportTable.isEmpty() {
		tmpl := template.Must(template.New("default.tmpl").ParseFS(content, "templates/default.tmpl"))
		type reportData struct {
			WarningMarker string
			ItemCount     int
			MainContent   string
		}

		processQueue := [][]*resourceData{
			reportTable.deleted,   // :red_circle:
			reportTable.created,   // :green_circle:
			reportTable.updated,   // :orange_circle:
			reportTable.unchanged, // :gray_circle:
		}

		for index, item := range processQueue {
			logger.Printf("Current index and item of ProcessQueue are: %d -> %v", index, item)

			if len(item) > 0 {
				tmplData := &reportData{
					ItemCount:   len(item),
					MainContent: formatMainContent(item).String(),
				}

				switch {
				case index == 0:
					tmplData.WarningMarker = ":red_circle:"
				case index == 1:
					tmplData.WarningMarker = ":green_circle:"
				case index == 2:
					tmplData.WarningMarker = ":orange_circle:"
				case index == 3:
					tmplData.WarningMarker = ":gray_circle:"
				}

				if err := tmpl.Execute(output, tmplData); err != nil {
					logger.Fatal(err)
				}
			}
		}
	} else {
		fmt.Fprint(output, "THERE IS NO ANY REPORT DATA")
	}
}

func tfPlanReader(pr *processingRequest) {
	logger.Printf("Process preparing for file: %s", pr.planPath)

	if pr.done {
		logger.Printf("Closing data channel after processing has finished for file: %s", pr.planPath)
		defer close(pr.parsedData)
	}

	logger.Printf("Waiting of green light in process pool for file: %s", pr.planPath)
	pr.pool <- 1
	logger.Printf("Green light has been acquired for file: %s", pr.planPath)

	if err := os.Chdir(path.Dir(pr.planPath)); err != nil {
		logger.Panicf("During changing current working directory to '%s', the error happened: %s", path.Dir(pr.planPath), err)
	}

	cmdResolvedPath, err := exec.LookPath(pr.commandName)
	if err != nil {
		log.Fatalf("It seems like the command %s can't be found", pr.commandName)
	}

	cmd := exec.Command(cmdResolvedPath, "show", "-json", "-no-color", path.Base(pr.planPath))
	var outputPlan strings.Builder
	cmd.Stdout = &outputPlan

	err = cmd.Run()
	if err != nil {
		cwd, _ := os.Getwd()
		log.Fatalf("Error happened during execution of command '%s' with CWD: %s: %s", cmd.String(), cwd, err)
	}

	var tfJsonPlan tfJson.Plan
	err = tfJsonPlan.UnmarshalJSON([]byte(outputPlan.String()))
	if err != nil {
		log.Fatalf("During unmarshalling TF json output the error happened: %s", err)
	}

	pr.parsedData <- tfJsonPlan

	logger.Printf("Processing was finished for file: %s", pr.planPath)
	//Return back capacity to the pool
	<-pr.pool
}
