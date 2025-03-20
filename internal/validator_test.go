package internal

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/arshvin/tf-plan-reporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SettingsValidatorTestSuite struct {
	suite.Suite

	tmpDir       string
	searchFolder string
	tfCmdFile    string

	settings *config.AppConfig
}

func (ts *SettingsValidatorTestSuite) SetupSuite() {
	ts.tmpDir = ts.T().TempDir() //nolint:typecheck
	ts.searchFolder = path.Join(ts.tmpDir, "some_folder")
	ts.tfCmdFile = path.Join(ts.tmpDir, "terraform")

	if _, err := os.Create(ts.tfCmdFile); err != nil {
		assert.FailNow(ts.T(), "Could not create file for test: %s", ts.tfCmdFile) //nolint:typecheck
	}

	if os.Mkdir(ts.searchFolder, 0750) != nil {
		assert.FailNow(ts.T(), "Could not create folder for test: %s", ts.searchFolder) //nolint:typecheck
	}
}

func (ts *SettingsValidatorTestSuite) SetupTest() {
	ts.settings = new(config.AppConfig)

	ts.settings.SearchFolder = ts.searchFolder
	ts.settings.TfCmdBinaryFile = ts.tfCmdFile
	ts.settings.TfPlanFileBasename = "tfplan.bin"
	ts.settings.CriticalResources = []string{"all"}
	ts.settings.IsAllCriticalSpecified = true
	ts.settings.AllowedRemovals = []string{"resource1", "resource2", "resource3"}
	ts.settings.NotUseTfChDirArg = false
}

func (ts *SettingsValidatorTestSuite) TestIfEmptyConfigParamHandled_terraform_binary_file() {
	ts.settings.TfCmdBinaryFile = ""
	err := Validate(ts.settings)

	errMsg := fmt.Sprintf(errMessageEmptyParam, "terraform_binary_file")
	assert.ErrorContains(ts.T(), err, errMsg, "Error message must be:  '%s'", errMsg) //nolint:typecheck
}
func (ts *SettingsValidatorTestSuite) TestIfEmptyConfigParamHandled_terraform_plan_file_basename() {
	ts.settings.TfPlanFileBasename = ""
	err := Validate(ts.settings)

	errMsg := fmt.Sprintf(errMessageEmptyParam, "terraform_plan_file_basename")
	assert.ErrorContains(ts.T(), err, errMsg, "Error message must be:  '%s'", errMsg) //nolint:typecheck
}
func (ts *SettingsValidatorTestSuite) TestIfEmptyConfigParamHandled_terraform_plan_search_folder() {
	ts.settings.SearchFolder = ""
	err := Validate(ts.settings)

	errMsg := fmt.Sprintf(errMessageEmptyParam, "terraform_plan_search_folder")
	assert.ErrorContains(ts.T(), err, errMsg, "Error message must be:  '%s'", errMsg) //nolint:typecheck
}

func (ts *SettingsValidatorTestSuite) TestIfAllShouldBeOnlyOne() {
	ts.settings.IsAllCriticalSpecified = true
	ts.settings.CriticalResources = append(ts.settings.CriticalResources, "resource1", "resource2")
	err := Validate(ts.settings)

	assert.ErrorContains(ts.T(), err, errMessageAllOnlyOne, "Error message must be:  '%s'", errMessageAllOnlyOne) //nolint:typecheck
}
func (ts *SettingsValidatorTestSuite) TestIfParamsCriticalAndAllowedFullBothHandled() {
	ts.settings.IsAllCriticalSpecified = false
	ts.settings.CriticalResources = []string{"resource1", "resource2"}
	err := Validate(ts.settings)

	assert.ErrorContains(ts.T(), err, errMessageCriticalAndAllowedFullBoth, "Error message must be:  '%s'", errMessageCriticalAndAllowedFullBoth) //nolint:typecheck
}
func (ts *SettingsValidatorTestSuite) TestIfParamsCriticalAndAllowedEmptyBothHandled() {
	ts.settings.IsAllCriticalSpecified = false
	ts.settings.CriticalResources = []string{}
	ts.settings.AllowedRemovals = []string{}
	err := Validate(ts.settings)

	assert.ErrorContains(ts.T(), err, errMessageCriticalAndAllowedEmptyBoth, "Error message must be:  '%s'", errMessageCriticalAndAllowedEmptyBoth) //nolint:typecheck

}
func (ts *SettingsValidatorTestSuite) TestIfAbsenceOfBinaryFileHandled() {
	ts.settings.TfCmdBinaryFile = ts.settings.TfCmdBinaryFile + "_absent"
	err := Validate(ts.settings) //Checking first case: if the path DOES NOT exist

	var pathError *os.PathError
	assert.ErrorAs(ts.T(), err, &pathError, "os.PathError must be be returned here") //nolint:typecheck

	ts.settings.TfCmdBinaryFile = ts.searchFolder
	err = Validate(ts.settings) //Checking second case: if the path DOES exist, but it's a folder
	errMsg := fmt.Sprintf(errMessagePathShouldNotBeFolder, ts.settings.TfCmdBinaryFile)
	assert.ErrorContains(ts.T(), err, errMsg, "Error message must be: '%s'", errMsg) //nolint:typecheck

}
func (ts *SettingsValidatorTestSuite) TestIfAbsenceOfSearchFolderHandled() {
	ts.settings.SearchFolder = ts.settings.SearchFolder + "_absent"
	err := Validate(ts.settings) //Checking first case: if the path DOES NOT exist

	var pathError *os.PathError
	assert.ErrorAs(ts.T(), err, &pathError, "os.PathError must be be returned here") //nolint:typecheck

	ts.settings.SearchFolder = ts.settings.TfCmdBinaryFile
	err = Validate(ts.settings) //Checking second case: if the path DOES exist, but it's a file
	errMsg := fmt.Sprintf(errMessagePathShouldNotBeFile, ts.settings.TfCmdBinaryFile)
	assert.ErrorContains(ts.T(), err, errMsg, "Error message must be: '%s'", errMsg) //nolint:typecheck
}
func (ts *SettingsValidatorTestSuite) TestIfTfProviderFolderAbsentHandled() {
	ts.settings.NotUseTfChDirArg = true
	err := Validate(ts.settings)

	assert.ErrorContains(ts.T(), err, errMessageTfProviderFolderAbsent, "Error message must be: '%s'", errMessageTfProviderFolderAbsent) //nolint:typecheck
}
func (ts *SettingsValidatorTestSuite) TestCorrectSettingsForTerraGRUNT1() {
	err := Validate(ts.settings)

	assert.Nil(ts.T(), err, "Application settings#1 must be valid") //nolint:typecheck
}
func (ts *SettingsValidatorTestSuite) TestCorrectSettingsForTerraGRUNT2() {
	ts.settings.IsAllCriticalSpecified = false
	ts.settings.CriticalResources = []string{"resource1", "resource2", "resource3"}
	ts.settings.AllowedRemovals = []string{}

	err := Validate(ts.settings)

	assert.Nil(ts.T(), err, "Application settings#2 must be valid") //nolint:typecheck
}
func (ts *SettingsValidatorTestSuite) TestCorrectSettingsForTerraFORM1() {
	ts.settings.NotUseTfChDirArg = true
	ts.settings.IsAllCriticalSpecified = false
	ts.settings.CriticalResources = []string{"resource1", "resource2", "resource3"}
	ts.settings.AllowedRemovals = []string{}

	ts.T().Chdir(ts.tmpDir) //nolint:typecheck
	if err := os.MkdirAll(".terraform/providers", 0755); err != nil {
		assert.FailNow(ts.T(), "Could not create testing folders: .terraform/providers") //nolint:typecheck

	}

	err := Validate(ts.settings)

	assert.Nil(ts.T(), err, "Application settings#3 must be valid") //nolint:typecheck
}

// Entry point for the test suite
func TestSettingsValidator(t *testing.T) {
	suite.Run(t, new(SettingsValidatorTestSuite))
}
