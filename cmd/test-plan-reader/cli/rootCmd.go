package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"

	tfJson "github.com/hashicorp/terraform-json"
	"github.com/mitchellh/cli"
)

type showCmd struct {
	HelpText     string
	SynopsisText string
}

func (c *showCmd) Help() string {
	return c.HelpText
}

func (c *showCmd) Run(args []string) int {

	flagSet:= flag.NewFlagSet(args[0],flag.ExitOnError)
	flagSet.Bool("json", true, "Outputs the plan in JSON format")
	flagSet.Bool("no-color", true, "Removes all color related pseudo symbols from output")

	if err:=flagSet.Parse(args); err != nil{
		logger.Fatalf("During parsing CLI args the error has happened: %s", err)
	}

	jsonPlanFileName := flagSet.Arg(0) //it should be json-file-with-plan
	inputFile, err := os.Open(jsonPlanFileName)
	if err != nil {
		logger.Fatalf("During opening file '%s' the error has happened: %s", jsonPlanFileName, err)
	}
	defer inputFile.Close()

	var planFilePath string
	if path.IsAbs((jsonPlanFileName)) {
		planFilePath = jsonPlanFileName
	} else {
		cwd, _ := os.Getwd()
		planFilePath = path.Join(cwd, jsonPlanFileName)
	}
	dirFS := os.DirFS(path.Dir(planFilePath))

	fileInfo, err := fs.Stat(dirFS, path.Base(planFilePath))
	if err != nil {
		logger.Fatalf("During gathering file info '%s' the error has happened: %s", planFilePath, err)
	}

	inputBuffer := make([]byte, fileInfo.Size())
	_, err = inputFile.Read(inputBuffer)
	if err != nil {
		logger.Fatalf("During reading file '%s' the error has happened: %s", planFilePath, err)
	}

	var fileContent tfJson.Plan
	if err = json.Unmarshal(inputBuffer, &fileContent); err != nil {
		logger.Fatalf("During decoding of file content '%s' the error has happened: %s", planFilePath, err)
	}

	outputBuffer, err :=json.MarshalIndent(fileContent,"","\t")
	if err != nil{
		logger.Fatalf("During preparing JSON to output the error has happened: %s", err)
	}

	fmt.Println(string(outputBuffer))

	return 0
}

func (c *showCmd) Synopsis() string {
	return c.SynopsisText
}

var (
	logger = log.Default()
)

func Execute() {

	appCli := cli.NewCLI("test-plan-reader", "v0.0.1")
	appCli.Args = os.Args[1:]
	appCli.Commands = map[string]cli.CommandFactory{
		"show": func() (cli.Command, error) {
			return &showCmd{
				SynopsisText: "This is the command which should mimic of behavior `show` argument of `terraform` command",
				HelpText: `Usage: test-plan-reader show [<args>] json-file-with-plan

Available flags are:
	-json    				Outputs the plan in JSON format
	-no-color				Removes all color related pseudo symbols from output
	json-file-with-plan		Filename with JSON formatted 'terraform plan' result
`,
			}, nil
		},
	}

	//nolint: errcheck
	appCli.Run()
}
