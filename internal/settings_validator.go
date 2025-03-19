package internal

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/arshvin/tf-plan-reporter/internal/config"
	"github.com/arshvin/tf-plan-reporter/internal/processing"
	log "github.com/sirupsen/logrus"
)

const (
	errMessageEmptyParam                  = "config file parameter should not be an empty string: '%s'"
	errMessageAllOnlyOne                  = "if config file parameter 'critical_resources' list contains 'all' value, its length must not be greater than 1 for preventing of ambiguity"
	errMessageCriticalAndAllowedFullBoth  = "if config file parameter 'critical_resources' list has some particular resources list, the 'allowed_removals' must be empty"
	errMessageCriticalAndAllowedEmptyBoth = "either config file parameter 'critical_resources' list or 'allowed_removals' list must be specified"
	errMessagePathShouldNotBeFolder       = "path should not be folder, but regular file instead: '%s'"
	errMessagePathShouldNotBeFile         = "path should not be regular file, but folder instead: '%s'"
	errMessageTfProviderFolderAbsent = "terraform providers folder (.terraform/providers) was not found in current working directory, which is mandatory if config file parameter 'not_use_chdir': true"
)

func Validate(settings *config.AppConfig) error {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get current working dir")
	}

	// First block of checks
	if err := errors.Join(
		checkIfParameterWasSpecified(settings.TfCmdBinaryFile, fmt.Sprintf(errMessageEmptyParam, "terraform_binary_file")),
		checkIfParameterWasSpecified(settings.TfPlanFileBasename, fmt.Sprintf(errMessageEmptyParam, "terraform_plan_file_basename")),
		checkIfParameterWasSpecified(settings.SearchFolder, fmt.Sprintf(errMessageEmptyParam, "terraform_plan_search_folder")),
		func() error {
			log.Debug("Checking if config file parameter 'critical_resources' is 'all' and there is only 1 item then ")

			if settings.IsAllCriticalSpecified && len(settings.CriticalResources) > 1 {
				return errors.New(errMessageAllOnlyOne)
			}

			return nil
		}(),
		func() error {
			log.Debug("Checking if config file parameter 'critical_resources' contains particular resources list then 'allowed_removals' must be empty")

			if !settings.IsAllCriticalSpecified && len(settings.CriticalResources) > 0 {
				if len(settings.AllowedRemovals) > 0 {
					return errors.New(errMessageCriticalAndAllowedFullBoth)
				}
			}

			return nil
		}(),
		func() error {
			log.Debug("Checking if config file parameters 'critical_resources' & 'allowed_removals' are not empty both")

			if len(settings.CriticalResources) == 0 && len(settings.AllowedRemovals) == 0 {
				return errors.New(errMessageCriticalAndAllowedEmptyBoth)
			}

			return nil
		}(),
	); err != nil {
		return err
	}

	//Replacing of relative TF command path to absolute one if it's required
	if !path.IsAbs(settings.TfCmdBinaryFile) {
		settings.TfCmdBinaryFile = path.Join(cwd, settings.TfCmdBinaryFile)
	}

	//Replacing of relative SearchFolder path to absolute one if it's required
	if !path.IsAbs(settings.SearchFolder) {
		settings.SearchFolder = path.Join(cwd, settings.SearchFolder)
	}

	// Second block of checks
	if err := errors.Join(
		checkIfPathExists(settings.TfCmdBinaryFile, true),
		checkIfPathExists(settings.SearchFolder, false),
		func() error { 	//Similar checking, if settings.NotUseTfChDirArg == false, will be further once all tf-plan files found
			if settings.NotUseTfChDirArg {
				log.Debug("Checking if Terraform providers folder exists in current folder in advance, 'not_use_chdir': true")

				if !processing.TfProviderFolderExist(cwd) {
					return errors.New(errMessageTfProviderFolderAbsent)
				}
			}

			return nil
		}(),
	); err != nil {
		return err
	}

	return nil
}

func checkIfParameterWasSpecified(parameterValue string, errMsg string) error {
	log.Debugf("Checking if config file parameter '%s' IS NOT empty string", parameterValue)

	if parameterValue == "" {
		return errors.New(errMsg)
	}

	return nil

}

func checkIfPathExists(filePath string, mustBeFile bool) error {
	log.Debugf("Checking if the file path '%s' exists", filePath)

	if stat, err := os.Stat(filePath); err != nil {
		return err
	} else {
		switch mustBeFile {
		case true:
			if stat.IsDir() {
				return fmt.Errorf(errMessagePathShouldNotBeFolder, filePath)
			}
		case false:
			if !stat.IsDir() {
				return fmt.Errorf(errMessagePathShouldNotBeFile, filePath)
			}
		}
	}

	return nil
}
