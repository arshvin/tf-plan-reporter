package processing

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"text/template"

	app "tf-plan-reporter/internal/config"

	tfJson "github.com/hashicorp/terraform-json"
)

var (
	logger = log.Default()
	table  *consolidatedJson

	//go:embed templates/*
	content embed.FS
)

func RunSearch(config app.ConfigFile) {
	searchFolder := os.DirFS(config.SearchFolder)
	tfPlanFilesPathList, err := fs.Glob(searchFolder, config.PlanFileBasename)
	if err != nil {
		log.Fatal(err)
	}

	dataPipe := make(chan tfJson.Plan)

	for _, tfPlanFilePath := range tfPlanFilesPathList {
		go tfPlanGetter(tfPlanFilePath, config.BinaryFile, dataPipe)
	}

	table = new(consolidatedJson)
	for tfPlan := range dataPipe {
		resourceChanges := tfPlan.ResourceChanges

		for _, resource := range resourceChanges {
			resourceItem := &resourceData{
				resourceType:  resource.Type,
				resourceName:  resource.Name,
				resourceIndex: resource.Index,
			}

			switch {
			case slices.Contains(resource.Change.Actions, tfJson.ActionCreate):
				table.created = append(table.created, resourceItem)

			case slices.Contains(resource.Change.Actions, tfJson.ActionDelete):
				table.deleted = append(table.deleted, resourceItem)

			case slices.Contains(resource.Change.Actions, tfJson.ActionUpdate):
				table.updated = append(table.updated, resourceItem)

			case slices.Contains(resource.Change.Actions, tfJson.ActionNoop):
				table.unchanged = append(table.unchanged, resourceItem)
			}
		}
	}
}

func PrintReport(filename string) {
	var output io.Writer

	switch {
	case len(filename) == 0:
		output = os.Stdout
	case len(filename) > 0:
		var err error
		if output, err = os.Create(filename); err != nil {
			log.Fatal(err)
		}

		defer (output.(io.Closer)).Close()
	}

	tmpl := template.Must(template.New("default").ParseFS(content, "default.tmpl"))
	type reportData struct {
		warning_marker string
		item_count     int
		main_content   string
	}

	processQueue := [][]*resourceData{
		table.deleted,   // :red_circle:
		table.created,   // :green_circle:
		table.updated,   // :orange_circle:
		table.unchanged, // :gray_circle:
	}

	for index, item := range processQueue {
		if len(item) > 0 {
			tmplData := &reportData{
				item_count:   len(item),
				main_content: prepareReport(item).String(),
			}

			switch {
			case index == 0:
				tmplData.warning_marker = ":red_circle:"
			case index == 1:
				tmplData.warning_marker = ":green_circle:"
			case index == 2:
				tmplData.warning_marker = ":orange_circle:"
			case index == 3:
				tmplData.warning_marker = ":gray_circle:"
			}

			tmpl.Execute(output, tmplData)
		}
	}
}

func tfPlanGetter(planBinaryFile string, cmdBinaryPath string, output chan tfJson.Plan) {
	if err := os.Chdir(path.Dir(planBinaryFile)); err != nil {
		logger.Panicf("During changing current working directory to '%s', the error happened: %s", path.Dir(planBinaryFile), err)
	}

	cmdResolvedPath, err := exec.LookPath(cmdBinaryPath)
	if err != nil {
		log.Fatalf("It seems like the command %s can't be found", cmdBinaryPath)
	}

	cmd := exec.Command(cmdResolvedPath, "show", "-json", "-no-color", planBinaryFile)
	var outputPlan strings.Builder
	cmd.Stdout = &outputPlan

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error happened during execution of command '%s': %s", cmd.String(), err)
	}

	var tfJsonPlan tfJson.Plan
	err = tfJsonPlan.UnmarshalJSON([]byte(outputPlan.String()))
	if err != nil {
		log.Fatalf("During unmarshalling TF json output the error happened: %s", err)
	}

	output <- tfJsonPlan
}
