package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

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

	var chDir string

	flagSet := flag.NewFlagSet(args[0], flag.ExitOnError)
	flagSet.Bool("json", true, "Outputs the plan in JSON format")
	flagSet.Bool("no-color", true, "Removes all color related pseudo symbols from output")
	flagSet.StringVar(&chDir, "chdir", "", "Switch to a different working directory before executing the given subcommand.")

	if err := flagSet.Parse(args); err != nil {
		log.Fatalf("During parsing CLI args the error happened: %s", err)
	}

	jsonPlanFileName := flagSet.Arg(0) //it should be `json-file-with-plan` cli arg

	planContext := log.WithField("plan_file_name", jsonPlanFileName)

	if len(chDir) > 0 {
		if err := os.Chdir(chDir); err != nil {
			planContext.Fatalf("During chdir operation the error happened: %s", err)
		}
	}

	inputFile, err := os.Open(jsonPlanFileName)
	if err != nil {
		planContext.Fatalf("During file opening the error happened: %s", err)
	}
	defer inputFile.Close()

	inputBuffer, err := io.ReadAll(inputFile)
	if err != nil {
		planContext.Fatalf("During file reading the error happened: %s", err)
	}

	var fileContent tfJson.Plan
	if err = json.Unmarshal(inputBuffer, &fileContent); err != nil {
		planContext.Fatalf("During content decoding the error happened: %s", err)
	}

	outputBuffer, err := json.MarshalIndent(fileContent, "", " ")
	if err != nil {
		planContext.Fatalf("During preparing JSON to output the error happened: %s", err)
	}

	fmt.Println(string(outputBuffer))

	return 0 //Stub, but if we'd have an error the app is failed earlier
}

func (c *showCmd) Synopsis() string {
	return c.SynopsisText
}

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
	-chdir=DIR				Switch to a different working directory before executing the
							given subcommand
	json-file-with-plan		Filename with JSON formatted 'terraform plan' result
`,
			}, nil
		},
	}

	//nolint: errcheck
	appCli.Run()
}
