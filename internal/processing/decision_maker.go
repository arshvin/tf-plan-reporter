package processing

import log "github.com/sirupsen/logrus"
import "github.com/arshvin/tf-plan-reporter/internal/config"

var (
	instance *DecisionMaker
)

type DecisionMaker struct {
	config *config.AppConfig
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

//TODO: implement test for this function to make sure that it works as expected
func (dm *DecisionMaker) IsAllowed(resourceType string) bool {
	if dm.config.IsAllCriticalSpecified {
		if _, ok := dm.config.IgnoreList[resourceType]; ok {
			log.Debugf("Resource type %s found in IgnoreList -> Allowed to delete", resourceType)
			return true
		}
		log.Debugf("Resource type %s WAS NOT found in IgnoreList -> Forbidden to delete", resourceType)
		dm.config.CriticalRemovalsFound = true
		return false
	}

	if _, ok := dm.config.RescueList[resourceType]; ok {
		log.Debugf("Resource type %s found in RescueList -> Forbidden to delete", resourceType)
		dm.config.CriticalRemovalsFound = true
		return false
	}
	log.Debugf("Resource type %s found in RescueList -> Allowed to delete", resourceType)
	return true
}
