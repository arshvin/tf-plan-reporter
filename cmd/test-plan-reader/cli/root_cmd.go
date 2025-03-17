package cli

import (
	"os"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/mitchellh/cli"
)

func Execute() {

	appCli := cli.NewCLI("test-plan-reader", "v0.0.1")
	appCli.Commands = map[string]cli.CommandFactory{
		"show": func() (cli.Command, error) {
			return &showCmd{
				SynopsisText: showCmdSynopsis,
				HelpText:     showCmdHelp,
			}, nil
		},
	}

	newCwd, newArgs, err := extractChdirOption(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid -chdir option: %s", err)
	}

	if newCwd != "" {
		err := os.Chdir(newCwd)
		if err != nil {
			log.Fatalf("Error handling -chdir option: %s", err)
		}
	}

	appCli.Args = newArgs
	appCli.HelpFunc = func(commands map[string]cli.CommandFactory) string {
		defaultHelp := cli.BasicHelpFunc(appCli.Name)(appCli.Commands)

		insertedLines := []string{
			"Global options (use these before the subcommand, if any):",
			"    -chdir=DIR    Switch to a different working directory before executing the given subcommand.",
			"",
		}

		modifiedHelp := slices.Insert(strings.Split(defaultHelp, "\n"), 2, insertedLines...)
		return strings.Join(modifiedHelp, "\n")
	}


	appCli.Run()
}
