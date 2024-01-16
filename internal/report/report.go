package report

import (
	"cmp"
	"embed"
	"fmt"
	"io"
	"slices"
	"strings"
	"text/template"

	"github.com/alexeyco/simpletable"
	exch "github.com/arshvin/tf-plan-reporter/internal"
	cfg "github.com/arshvin/tf-plan-reporter/internal/config"
	log "github.com/sirupsen/logrus"
)

var (
	//go:embed templates
	content embed.FS
)

type reportData struct {
	WarningMarker string
	Caption       string
	ItemCount     int
	MainContent   string
}

type Report struct {
	template string
	output   io.Writer
	data     []*reportData
}

func (r *Report) Prepare(data *exch.ConsolidatedJson) {

	processQueue := map[string][]*exch.ResourceData{
		"delete": data.Deleted,
		"create": data.Created,
		"update": data.Updated,
		"noop":   data.Unchanged,
	}

	headers := map[string][]string{
		"delete": []string{":red_circle:", "Resources to be deleted"},
		"create": []string{":green_circle:", "Resources to be created"},
		"update": []string{":orange_circle:", "Resources to be updated"},
		"noop":   []string{":white_circle:", "Resources to be ignored for change"},
	}

	for key, value := range processQueue {

		tableContext := log.WithFields(log.Fields{
			"action_type":        key,
			"affected_resources": len(value),
		})
		tableContext.Debug("Statistics")

		if len(value) > 0 {

			tableContext.Debug("Preparing of resource table")

			item := &reportData{
				ItemCount:     len(value),
				MainContent:   formatMainContent(value, key == "delete").String(),
				WarningMarker: headers[key][0],
				Caption:       headers[key][1],
			}

			r.data = append(r.data, item)
		}
	}

}

func (r *Report) Print() {
	log.Debug("Output of resource table")

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
	}

	tmpl := template.Must(template.New("default.tmpl").Funcs(funcMap).ParseFS(content, "templates/default.tmpl"))

	for _, item := range r.data {

		if err := tmpl.Execute(r.output, item); err != nil {
			log.Fatal(err)
		}
	}

}

func ForGitHub(output io.Writer) *Report {
	return &Report{output: output, template: "templates/default.tmpl"}
}

func formatMainContent(items []*exch.ResourceData, isDeleteTable bool) *simpletable.Table {
	headers := []string{"Type", "Index", "Name"}

	if isDeleteTable {
		headers = slices.Insert(headers, 0, "Allowed to remove")
	}

	log.Debug("Instantiating of report table")
	table := simpletable.New()
	table.SetStyle(simpletable.StyleMarkdown)
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{},
	}

	log.Debug("Filling of report table headers")
	for _, header := range headers {
		table.Header.Cells = append(
			table.Header.Cells, &simpletable.Cell{Align: simpletable.AlignCenter, Text: header},
		)
	}

	log.Debug("Sorting elements data elements before table report filling")
	slices.SortFunc(items, func(a, b *exch.ResourceData) int {
		return cmp.Compare(a.Type, b.Type)
	})

	log.Debug("Filling of report table rows")
	for _, item := range items {
		row := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: item.Type},
			{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s", item.Index)},
			{Align: simpletable.AlignLeft, Text: item.Name},
		}

		if isDeleteTable {
			row = slices.Insert(row, 0,
				&simpletable.Cell{Align: simpletable.AlignCenter, Text: isAllowedToRemove(item.Type)})
		}

		table.Body.Cells = append(table.Body.Cells, row)
	}

	return table
}

func isAllowedToRemove(resourceType string) string {
	if cfg.AppConfig.IsAllCriticalSpecified {
		if _, ok := cfg.AppConfig.IgnoreList[resourceType]; ok {
			return ":white_check_mark: "
		}
		cfg.AppConfig.CriticalRemovalsFound = true
		return ":x:"
	} else {
		if _, ok := cfg.AppConfig.RescueList[resourceType]; ok {
			cfg.AppConfig.CriticalRemovalsFound = true
			return ":x:"
		}
		return ":white_check_mark: "
	}
}
