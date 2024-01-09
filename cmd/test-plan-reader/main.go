package main

import (
	"github.com/arshvin/tf-plan-reporter/cmd/test-plan-reader/cli"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		PadLevelText:           true,
		DisableTimestamp:       false,
		FullTimestamp:          true,
		ForceColors:            true,
	})
}

func main() {
	cli.Execute()
}
