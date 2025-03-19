package processing

import (
	"testing"

	"github.com/arshvin/tf-plan-reporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DecisionTestSuite struct {
	suite.Suite

	settings *config.AppConfig
}

func (ts *DecisionTestSuite) SetupSuite() {

	ts.settings = new(config.AppConfig)
	ts.settings.ExceptionalResources = make(map[string]bool)

	for _, item := range []string{"resource1", "resource2", "resource3"} {
		ts.settings.ExceptionalResources[item] = true
	}
}

func (ts *DecisionTestSuite) TestIfDeletingIsAllowed() {
	ts.settings.IsAllCriticalSpecified = true
	dm := GetDecisionMaker()
	dm.SetConfig(ts.settings)

	assert.True(ts.T(), dm.IsAllowedForRemoval("resource1"))  //nolint:typecheck
	assert.True(ts.T(), dm.IsAllowedForRemoval("resource2"))  //nolint:typecheck
	assert.True(ts.T(), dm.IsAllowedForRemoval("resource3"))  //nolint:typecheck
	assert.False(ts.T(), dm.IsAllowedForRemoval("resource4")) //nolint:typecheck
}

func (ts *DecisionTestSuite) TestIfDeletingIsForbidden() {
	ts.settings.IsAllCriticalSpecified = false
	dm := GetDecisionMaker()
	dm.SetConfig(ts.settings)

	assert.False(ts.T(), dm.IsAllowedForRemoval("resource1")) //nolint:typecheck
	assert.False(ts.T(), dm.IsAllowedForRemoval("resource2")) //nolint:typecheck
	assert.False(ts.T(), dm.IsAllowedForRemoval("resource3")) //nolint:typecheck
	assert.True(ts.T(), dm.IsAllowedForRemoval("resource4"))  //nolint:typecheck
}

// Entry point for the test suite
func TestSettingsValidator(t *testing.T) {
	suite.Run(t, new(DecisionTestSuite))
}
