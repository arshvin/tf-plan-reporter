package report

import (
	"cmp"
	"embed"
	"fmt"
	"io"
	"os"
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
	TableContent *simpletable.Table
	ActionType   byte
}

type Report struct {
	template   string
	output     io.Writer
	data       []*reportData
	answers    map[bool]string
	tableStyle *simpletable.Style
}

func ForGitHub(output io.Writer) *Report {
	return &Report{
		output:   output,
		template: "templates/github_markdown.tmpl",
		answers: map[bool]string{
			true:  ":white_check_mark:", // https://emojipedia.org/check-mark-button#technical
			false: ":x:",                // https://emojipedia.org/cross-mark#technical
		},
		tableStyle: simpletable.StyleMarkdown,
	}
}

func ForStdout() *Report {
	return &Report{
		output:   os.Stdout,
		template: "templates/stdout.tmpl",
		answers: map[bool]string{
			true:  "yes",
			false: "no",
		},
		tableStyle: simpletable.StyleUnicode,
	}
}

func (r *Report) Prepare(data *exch.ConsolidatedJson) {

	queue := []byte{deleted, created, updated, unchanged}

	var answers map[bool]string

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
			tableLogger := log.WithFields(
				log.Fields{
					"action_type":     actionType,
					"output_template": path.Base(r.template),
				})

			tableLogger.Debug("Preparing of main content of template section")

			if actionType == deleted {
				answers = r.answers
			} else {
				answers = nil
			}

			item := &reportData{
				TableContent: formatMainContent(r.tableStyle, value, answers, tableLogger),
				ItemCount:    amount,
				ActionType:   actionType,
			}

			r.data = append(r.data, item)
		}
	}
}

func (r *Report) Print() {

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
	}

	parentTemplate := template.Must(template.New(path.Base(r.template)).Funcs(funcMap).ParseFS(content, r.template))

	for _, item := range r.data {

		log.WithFields(
			log.Fields{
				"action_type":     item.ActionType,
				"output_template": path.Base(r.template),
			}).Debug("Output of result table")

		var templatePathName string

		switch item.ActionType {
		case deleted:
			templatePathName = r.getTemplate("deleted.tmpl")
		case created:
			templatePathName = r.getTemplate("created.tmpl")
		case updated:
			templatePathName = r.getTemplate("updated.tmpl")
		case unchanged:
			templatePathName = r.getTemplate("unchanged.tmpl")
		}

		resultTemplate := template.Must(template.Must(parentTemplate.Clone()).ParseFS(content, templatePathName))

		if err := resultTemplate.Execute(r.output, item); err != nil {
			log.Fatal(err)
		}
	}

}

func (r *Report) getTemplate(name string) string {
	return fmt.Sprintf("%s/%s", strings.Split(r.template, ".")[0], name)
}

func formatMainContent(tableStyle *simpletable.Style, items []*exch.ResourceData, deleteTableAnswers map[bool]string, logger *log.Entry) *simpletable.Table {
	headers := []string{"Type", "Name", "Index (if any)"}

	if deleteTableAnswers != nil {
		headers = slices.Insert(headers, 0, "Allowed to remove")
	}

	logger.Debug("Instantiating of report table")
	table := simpletable.New()
	table.SetStyle(tableStyle)
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{},
	}

	logger.Debug("Filling of report table headers")
	for _, header := range headers {
		table.Header.Cells = append(
			table.Header.Cells, &simpletable.Cell{Align: simpletable.AlignCenter, Text: header},
		)
	}

	logger.Debug("Sorting elements data elements before table report filling")
	slices.SortFunc(items, func(a, b *exch.ResourceData) int {
		return cmp.Compare(a.Type, b.Type)
	})

	logger.Debug("Filling of report table rows")
	for _, item := range items {
		row := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: item.Type},
			{Align: simpletable.AlignLeft, Text: item.Name},
			{Align: simpletable.AlignLeft, Text: item.Index},
		}

		if deleteTableAnswers != nil {
			answer := isAllowedToRemove(item.Type, deleteTableAnswers)
			logger.WithField("resource_type", item.Type).Debugf("Is it OK to remove: %s", answer)

			row = slices.Insert(row, 0,
				&simpletable.Cell{Align: simpletable.AlignCenter, Text: answer})
		}

		table.Body.Cells = append(table.Body.Cells, row)
	}

	return table
}

func isAllowedToRemove(resourceType string, deleteTableAnswers map[bool]string) string {
	if cfg.AppConfig.IsAllCriticalSpecified {
		if _, ok := cfg.AppConfig.IgnoreList[resourceType]; ok {
			return deleteTableAnswers[true]
		}
		cfg.AppConfig.CriticalRemovalsFound = true
		return deleteTableAnswers[false]
	} else {
		if _, ok := cfg.AppConfig.RescueList[resourceType]; ok {
			cfg.AppConfig.CriticalRemovalsFound = true
			return deleteTableAnswers[false]
		}
		return deleteTableAnswers[true]
	}
}
