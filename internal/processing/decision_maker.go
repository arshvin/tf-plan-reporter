package processing

import (
	"os"
	"path"

	"github.com/arshvin/tf-plan-reporter/internal/config"
	log "github.com/sirupsen/logrus"
)

var (
	instance *DecisionMaker
)

const (
	tfProvidersFolderName = ".terraform/providers"
)

type DecisionMaker struct {
	config                *config.AppConfig
	criticalRemovalsFound bool
	IsAllowedForRemoval   func(string) bool
}

func GetDecisionMaker() *DecisionMaker {
	if instance == nil {
		instance = new(DecisionMaker)
	}

	return instance
}

func (dm *DecisionMaker) SetConfig(config *config.AppConfig) {
	dm.config = config

	if dm.config.IsAllCriticalSpecified {
		dm.IsAllowedForRemoval = dm.isInAllowedList
	} else {
		dm.IsAllowedForRemoval = dm.isInRescueList
	}

}

func (dm *DecisionMaker) CriticalRemovalsFound() bool {
	return dm.criticalRemovalsFound
}

func (dm *DecisionMaker) isInAllowedList(resourceType string) bool {
	if _, ok := dm.config.ExceptionalResources[resourceType]; ok {
		log.Debugf("Resource type %s found in IgnoreList -> Allowed to delete", resourceType)

		return true
	}

	log.Debugf("Resource type %s WAS NOT found in IgnoreList -> Forbidden to delete", resourceType)
	dm.criticalRemovalsFound = true

	return false

}

func (dm *DecisionMaker) isInRescueList(resourceType string) bool {
	if _, ok := dm.config.ExceptionalResources[resourceType]; ok {
		log.Debugf("Resource type %s found in RescueList -> Forbidden to delete", resourceType)
		dm.criticalRemovalsFound = true

		return false
	}

	log.Debugf("Resource type %s found in RescueList -> Allowed to delete", resourceType)

	return true

}

func TfProviderFolderExist(prefix string) bool {
	providersFolder := path.Join(prefix, tfProvidersFolderName)
	info, err := os.Lstat(providersFolder)

	if err != nil {
		log.Debugf("Provider folder does not exist: %s", providersFolder)
		return false
	}

	if info.IsDir() {
		log.Debugf("Provider folder was successfully found: %s", providersFolder)

		return true
	}

	log.Debugf("The path is not folder: %s", providersFolder)

	return false
}
