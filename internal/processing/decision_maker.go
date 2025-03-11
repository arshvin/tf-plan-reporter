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
	config *config.AppConfig
	criticalRemovalsFound  bool

}

func GetDecisionMaker() *DecisionMaker {
	if instance == nil {
		instance = new(DecisionMaker)
	}

	return instance
}

func (dm *DecisionMaker) SetConfig(config *config.AppConfig) {
	dm.config = config
}

func (dm *DecisionMaker) CriticalRemovalsFound() bool{
	return dm.criticalRemovalsFound
}

//TODO: implement test for this function to make sure that it works as expected
func (dm *DecisionMaker) IsAllowed(resourceType string) bool {
	if dm.config.IsAllCriticalSpecified {
		if _, ok := dm.config.IgnoreList[resourceType]; ok {
			log.Debugf("Resource type %s found in IgnoreList -> Allowed to delete", resourceType)
			return true
		}

		log.Debugf("Resource type %s WAS NOT found in IgnoreList -> Forbidden to delete", resourceType)
		dm.criticalRemovalsFound = true

		return false
	}

	if _, ok := dm.config.RescueList[resourceType]; ok {
		log.Debugf("Resource type %s found in RescueList -> Forbidden to delete", resourceType)
		dm.criticalRemovalsFound = true

		return false
	}

	log.Debugf("Resource type %s found in RescueList -> Allowed to delete", resourceType)

	return true
}

func TfProviderFolderExist(prefix string) bool {
	providersFolder :=path.Join(prefix,tfProvidersFolderName)
	info, err := os.Lstat(providersFolder)

	if err != nil {
		log.Fatalf("An error has happened: %s",err)
	}

	if info.IsDir() {
		log.Debugf("Provider folder was successfully FOUND: %s", providersFolder)

		return true
	}

	log.Debugf("Provider folder WAS NOT found: %s", providersFolder)

	return false
}
