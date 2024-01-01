package main

import (
	"tf-plan-reporter/cmd/tf-plan-reporter/cli"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		PadLevelText:           true,
		DisableTimestamp:       false,
		FullTimestamp:          true,
	})
}

func main() {
	cli.Execute()
}
