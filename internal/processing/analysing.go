package processing

import (
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	app "tf-plan-reporter/internal/config"

	tfJson "github.com/hashicorp/terraform-json"
)

var(
	logger = log.Default()
)

func RunSearch(config app.ConfigFile) {
	searchFolder := os.DirFS(config.SearchFolder)
	tfPlanFilesPathList, err := fs.Glob(searchFolder, config.PlanFileBasename)
	if err != nil {
		log.Fatal(err)
	}
	
	dataPipe := make(chan tfJson.Plan)

	for _,tfPlanFilePath := range tfPlanFilesPathList {
		go tfPlanGetter(tfPlanFilePath,config.BinaryFile,dataPipe)	
	}

	table
	for tfPlan := dataPipe{

	}
}

func tfPlanGetter(planBinaryFile string, cmdBinaryPath string, output chan tfJson.Plan){
	if err:=os.Chdir(path.Dir(planBinaryFile)); err!=nil {
		logger.Panicf("During changing current working directory to '%s', the error happened: %s",path.Dir(planBinaryFile), err)
	}

	cmdResolvedPath, err := exec.LookPath(cmdBinaryPath)
	if err != nil {
		log.Fatalf("It seems like the command %s can't be found", cmdBinaryPath)
	}

	cmd :=exec.Command(cmdResolvedPath, "show", "-json", "-no-color", planBinaryFile)
	var outputPlan strings.Builder
	cmd.Stdout = &outputPlan
	
	err = cmd.Run() 
	if err != nil{
		log.Fatalf("Error happened during execution of command '%s': %s", cmd.String(), err)
	}
	
	var tfJsonPlan tfJson.Plan
	err = tfJsonPlan.UnmarshalJSON([]byte(outputPlan.String()))
	if err != nil{
		log.Fatalf("During unmarshalling TF json output the error happened: %s", err)
	}

	output <- tfJsonPlan
}

