package report

import (
	"cmp"
	"embed"
	"io"
	"path"
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

const (
	deleted byte = iota
	created
	updated
	unchanged
)

type reportData struct {
	ItemCount    int
	TableContent string
	ActionType   byte
}

type Report struct {
	template string
	output   io.Writer
	data     []*reportData
}

func (r *Report) Prepare(data *exch.ConsolidatedJson) {

	queue := []byte{deleted, created, updated, unchanged}

	for _, actionType := range queue {
		var value []*exch.ResourceData

		switch actionType {
		case deleted:
			value = data.Deleted
		case created:
			value = data.Created
		case updated:
			value = data.Updated
		case unchanged:
			value = data.Unchanged
		}

		amount := len(value)

		if amount > 0 {

			log.Debug("Preparing of main content of template section")

			item := &reportData{
				TableContent: formatMainContent(value, actionType == deleted).String(),
				ItemCount:    amount,
				ActionType:   actionType,
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

	parentTemplate := template.Must(template.New(path.Base(r.template)).Funcs(funcMap).ParseFS(content, r.template))

	for _, item := range r.data {
		var templatePathName string

		switch item.ActionType {
		case deleted:
			templatePathName = "templates/github_markdown/deleted.tmpl"
		case created:
			templatePathName = "templates/github_markdown/created.tmpl"
		case updated:
			templatePathName = "templates/github_markdown/updated.tmpl"
		case unchanged:
			templatePathName = "templates/github_markdown/unchanged.tmpl"
		}

		resultTemplate := template.Must(template.Must(parentTemplate.Clone()).ParseFS(content, templatePathName))

		if err := resultTemplate.Execute(r.output, item); err != nil {
			log.Fatal(err)
		}
	}

}

func ForGitHub(output io.Writer) *Report {
	return &Report{output: output, template: "templates/github_markdown.tmpl"}
}

func formatMainContent(items []*exch.ResourceData, isDeleteTable bool) *simpletable.Table {
	headers := []string{"Type", "Name", "Index (if any)"}

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
			{Align: simpletable.AlignLeft, Text: item.Name},
			{Align: simpletable.AlignLeft, Text: item.Index},
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
