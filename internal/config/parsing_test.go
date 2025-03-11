package config

import (
	"path"
	"testing"

	"os"

	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/assert"
)

type ConfigParserTestSuite struct {
	suite.Suite
	tmpDir string
}

func (ts *ConfigParserTestSuite) SetupSuite() {
	ts.tmpDir = ts.T().TempDir()
}

func (ts *ConfigParserTestSuite) TestParsing1() {
	fileContent := `
terraform_binary_file: /usr/local/bin/terraform
terraform_plan_file_basename: tfplan.bin
terraform_plan_search_folder: ./
`

	fileName := path.Join(ts.tmpDir, "config1.yaml")

	configFile, err := os.Create(fileName)
	if err != nil {
		ts.T().Fatalf("Could not create tmp file: %s", fileName)
	}

	_, err = configFile.WriteString(fileContent)
	if err != nil {
		ts.T().Fatalf("Could not write file content: %s", fileName)
	}

	err = configFile.Close()
	if err != nil {
		ts.T().Fatalf("Could save file: %s", fileName)
	}

	parsedConfig := Parse(fileName)

	cwd,_:=os.Getwd()

	assert.Equal(ts.T(), "/usr/local/bin/terraform", parsedConfig.TfCmdBinaryFile)
	assert.Equal(ts.T(), "tfplan.bin", parsedConfig.TfPlanFileBasename)
	assert.Equal(ts.T(), cwd, parsedConfig.SearchFolder)
	assert.Equal(ts.T(), false, parsedConfig.IsAllCriticalSpecified)
	assert.Equal(ts.T(), []string(nil), parsedConfig.AllowedRemovals)
	assert.Equal(ts.T(), []string(nil), parsedConfig.CriticalResources)
	assert.Equal(ts.T(), false, parsedConfig.NotUseTfChDirArg)

}

func TestConfigParsingDefault(t *testing.T) {
	suite.Run(t, new(ConfigParserTestSuite))
}
