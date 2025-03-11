package processing

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	tfJson "github.com/hashicorp/terraform-json"
)

type processingRequest struct {
	commandName string
	planPath    string
	parsedData  chan<- tfJson.Plan
	pool        chan int
	notChDir    bool
}

// CollectBinaryData function does:
// 1. searches all terraform generated binary plan files, with basename specified in `terraform_plan_file_basename`,
// starting from root director specified in `terraform_plan_search_folder` config file parameter
// 2. fills of `reportData` variable, by parsed terraform plan data, which further is going to be source of printed report
func CollectBinaryData(startPath string, planBaseFileName string, cmdFullPathName string, notChDir bool) *ConsolidatedJson {
	foundPlanFiles := findAllTFPlanFiles(startPath, planBaseFileName)

	foundItems := len(foundPlanFiles)
	log.WithFields(log.Fields{
		"plan_basename": planBaseFileName,
		"total_amount":  foundItems,
	}).Debug("Found terraform generated plan files")

	reportData := new(ConsolidatedJson)

	if foundItems > 0 {

		pool := make(chan int, runtime.GOMAXPROCS(0))
		dataPipe := make(chan tfJson.Plan, runtime.GOMAXPROCS(0))

		absCmdBinaryPath := cmdFullPathName
		if !path.IsAbs(cmdFullPathName) { //TODO:This is not a goal of this function - it should give what it has been passed to and use. It'd be better to implement some "validation" config step outside of this function
			cwd, _ := os.Getwd()
			absCmdBinaryPath = path.Join(cwd, cmdFullPathName)
		}

		for _, absTFPlanFilePath := range foundPlanFiles {
			pr := &processingRequest{
				commandName: absCmdBinaryPath,
				planPath:    absTFPlanFilePath,
				parsedData:  dataPipe,
				pool:        pool,
				notChDir:    notChDir,
			}

			go tfPlanReader(pr) //Async TF plan reader
		}

		//Parsing of read data
		log.Debug("Waiting of data from read TF plan files for processing")
		for item := 0; item < foundItems; item++ {
			tfPlan := <-dataPipe

			reportData.Parse(&tfPlan)
		}
	}

	return reportData
}

//TODO: Implement test of this function to make sure that it works as expected
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

func tfPlanReader(pr *processingRequest) {
	planFileContext := log.WithField("plan_file_name", pr.planPath)
	planFileContext.Info("Preparation for parsing")

	planFileContext.Debug("Waiting of green light in process pool")
	pr.pool <- 1
	planFileContext.Debug("Green light has been acquired")

	cmdResolvedPath, err := exec.LookPath(pr.commandName)
	if err != nil {
		planFileContext.Fatalf("Could not find the command file: %s", pr.commandName)
	}

	var auxCmdArgs string
	if pr.notChDir { //TODO: To leave couple lines of comments here about each case: what and why is that?
		auxCmdArgs = fmt.Sprintf("show -json -no-color %s", pr.planPath)
	} else {
		planDirName := path.Dir(pr.planPath)
		auxCmdArgs = fmt.Sprintf("-chdir=%s show -json -no-color %s", planDirName, path.Base(pr.planPath))
	}

	cmdContext := planFileContext.WithFields(log.Fields{
		"command":        cmdResolvedPath,
		"args":           auxCmdArgs,
		"plan_file_name": path.Base(pr.planPath),
	})
	cmdContext.Debug("Command launching")

	cmd := exec.Command(cmdResolvedPath, strings.Split(auxCmdArgs, " ")...)
	var outputPlan strings.Builder
	var tfErr strings.Builder

	cmd.Stdout = &outputPlan
	cmd.Stderr = &tfErr

	err = cmd.Run()
	if err != nil {
		cmdContext.Debugf("Command stderr output:\n%s", tfErr.String())
		cmdContext.Fatalf("During execution the error happened: %s", err)
	}

	var tfJsonPlan tfJson.Plan
	err = tfJsonPlan.UnmarshalJSON([]byte(outputPlan.String()))
	if err != nil {
		planFileContext.Fatalf("Could not unmarshal: %s", err)
	}

	planFileContext.Debugf("Harvested records: %v", len(tfJsonPlan.ResourceChanges))

	pr.parsedData <- tfJsonPlan

	planFileContext.Print("Parsing finished")
	//Return back capacity to the pool
	<-pr.pool
}
