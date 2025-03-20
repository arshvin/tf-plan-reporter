package config

import (
	"path"
	"testing"

	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigParserTestSuite struct {
	suite.Suite
	tmpDir string
}

func (ts *ConfigParserTestSuite) SetupSuite() {
	ts.tmpDir = ts.T().TempDir()
}

func (ts *ConfigParserTestSuite) createFile(name string, content string) {
	configFile, err := os.Create(name)
	if err != nil {
		ts.T().Fatalf("Could not create tmp file: %s", name) //nolint:typecheck
	}

	_, err = configFile.WriteString(content)
	if err != nil {
		ts.T().Fatalf("Could not write file content: %s", name) //nolint:typecheck
	}

	err = configFile.Close()
	if err != nil {
		ts.T().Fatalf("Could save file: %s", name) //nolint:typecheck
	}

}
func (ts *ConfigParserTestSuite) TestParsingSomeMinimalConfig() {
	fileContent := `
terraform_binary_file: /usr/local/bin/terraform
terraform_plan_file_basename: tfplan.bin
terraform_plan_search_folder: ./
`

	fileName := path.Join(ts.tmpDir, "config_minimal.yaml")

	ts.createFile(fileName, fileContent)

	parsedConfig := Parse(fileName)

	assert.Equal(ts.T(), "/usr/local/bin/terraform", parsedConfig.TfCmdBinaryFile) //nolint:typecheck
	assert.Equal(ts.T(), "tfplan.bin", parsedConfig.TfPlanFileBasename)            //nolint:typecheck
	assert.Equal(ts.T(), "./", parsedConfig.SearchFolder)                          //nolint:typecheck
	assert.Equal(ts.T(), false, parsedConfig.IsAllCriticalSpecified)               //nolint:typecheck
	assert.Equal(ts.T(), []string(nil), parsedConfig.AllowedRemovals)              //nolint:typecheck
	assert.Equal(ts.T(), []string(nil), parsedConfig.CriticalResources)            //nolint:typecheck
	assert.Equal(ts.T(), false, parsedConfig.NotUseTfChDirArg)                     //nolint:typecheck
	assert.Equal(ts.T(), 0, len(parsedConfig.ExceptionalResources))                //nolint:typecheck

}

func (ts *ConfigParserTestSuite) TestParsingEmptyFile() {
	fileContent := ""

	fileName := path.Join(ts.tmpDir, "config_empty.yaml")

	ts.createFile(fileName, fileContent)

	parsedConfig := Parse(fileName)

	assert.Equal(ts.T(), "", parsedConfig.TfCmdBinaryFile)              //nolint:typecheck
	assert.Equal(ts.T(), "", parsedConfig.TfPlanFileBasename)           //nolint:typecheck
	assert.Equal(ts.T(), "", parsedConfig.SearchFolder)                 //nolint:typecheck
	assert.Equal(ts.T(), false, parsedConfig.IsAllCriticalSpecified)    //nolint:typecheck
	assert.Equal(ts.T(), []string(nil), parsedConfig.AllowedRemovals)   //nolint:typecheck
	assert.Equal(ts.T(), []string(nil), parsedConfig.CriticalResources) //nolint:typecheck
	assert.Equal(ts.T(), false, parsedConfig.NotUseTfChDirArg)          //nolint:typecheck
	assert.Equal(ts.T(), 0, len(parsedConfig.ExceptionalResources))     //nolint:typecheck

}

func (ts *ConfigParserTestSuite) TestParsingFullyFilledFile() {
	fileContent := `
terraform_binary_file: ./test-plan-reader
terraform_plan_file_basename: plan.json
terraform_plan_search_folder: /tmp
not_use_chdir: true


critical_resources:
  - all

allowed_removals:  # makes sense only if "all" specified in "critical_resources" section
  - resource_type1
  - resource_type2
  - resource_type3
`

	fileName := path.Join(ts.tmpDir, "config_full.yaml")

	ts.createFile(fileName, fileContent)

	parsedConfig := Parse(fileName)

	assert.Equal(ts.T(), "./test-plan-reader", parsedConfig.TfCmdBinaryFile)                                           //nolint:typecheck
	assert.Equal(ts.T(), "plan.json", parsedConfig.TfPlanFileBasename)                                                 //nolint:typecheck
	assert.Equal(ts.T(), "/tmp", parsedConfig.SearchFolder)                                                            //nolint:typecheck
	assert.Equal(ts.T(), true, parsedConfig.IsAllCriticalSpecified)                                                    //nolint:typecheck
	assert.Equal(ts.T(), []string{"resource_type1", "resource_type2", "resource_type3"}, parsedConfig.AllowedRemovals) //nolint:typecheck
	assert.Equal(ts.T(), []string{"all"}, parsedConfig.CriticalResources)                                              //nolint:typecheck
	assert.Equal(ts.T(), true, parsedConfig.NotUseTfChDirArg)                                                          //nolint:typecheck
	assert.Equal(ts.T(), 3, len(parsedConfig.ExceptionalResources))                                                    //nolint:typecheck

}

// Entry point for the test suite
func TestConfigParsingDefault(t *testing.T) {
	suite.Run(t, new(ConfigParserTestSuite))
}
