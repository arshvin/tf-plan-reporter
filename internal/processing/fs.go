package processing

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	tfJson "github.com/hashicorp/terraform-json"
	log "github.com/sirupsen/logrus"
)

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
	planFileContext.Printf("Preparation for parsing")

	planFileContext.Debug("Waiting of green light in process pool")
	pr.pool <- 1
	planFileContext.Debug("Green light has been acquired")

	cmdResolvedPath, err := exec.LookPath(pr.commandName)
	if err != nil {
		planFileContext.Fatalf("Could not find the command file: %s", pr.commandName)
	}

	cmdContext := planFileContext.WithFields(log.Fields{
		"command":        cmdResolvedPath,
		"plan_file_name": path.Base(pr.planPath),
	})
	cmdContext.Debug("Launching of command")

	planDirName := path.Dir(pr.planPath)
	cmd := exec.Command(cmdResolvedPath, "show", fmt.Sprintf("-chdir=%s", planDirName), "-json", "-no-color", path.Base(pr.planPath))
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

	planFileContext.Debugf("Harvestered records: %v", len(tfJsonPlan.ResourceChanges))

	pr.parsedData <- tfJsonPlan

	planFileContext.Print("Parsing finished")
	//Return back capacity to the pool
	<-pr.pool
}
