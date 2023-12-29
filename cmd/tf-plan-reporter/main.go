package main

import (
	"tf-plan-reporter/cmd/tf-plan-reporter/cli"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		PadLevelText:           true,
		DisableTimestamp:       false,
		FullTimestamp:          true,
		ForceColors:            false,
	})
}

func main() {
	cli.Execute()
}
