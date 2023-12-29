package processing

import (
	"cmp"
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"

	cfg "tf-plan-reporter/internal/config"

	"github.com/alexeyco/simpletable"
)

type resourceData struct {
	resourceType  string
	resourceName  string
	resourceIndex interface{}
}

type consolidatedJson struct {
	created   []*resourceData
	updated   []*resourceData
	deleted   []*resourceData
	unchanged []*resourceData
}

func (cj *consolidatedJson) isEmpty() bool {
	return len(cj.created)+len(cj.updated)+len(cj.deleted)+len(cj.unchanged) == 0
}

func (cj *consolidatedJson) totalItems() int {
	return len(cj.created) + len(cj.updated) + len(cj.deleted) + len(cj.unchanged)
}

func formatMainContent(items []*resourceData, isDeleteTable bool) *simpletable.Table {
	headers := []string{"Type", "Index", "Name"}

	if isDeleteTable {
		headers = slices.Insert(headers, 0, "Allowable to remove")
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
	slices.SortFunc(items, func(a, b *resourceData) int {
		return cmp.Compare(a.resourceType, b.resourceType)
	})

	log.Debug("Filling of report table rows")
	for _, item := range items {
		row := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: item.resourceType},
			{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s", item.resourceIndex)},
			{Align: simpletable.AlignLeft, Text: item.resourceName},
		}

		if isDeleteTable {
			row = slices.Insert(row, 0,
				&simpletable.Cell{Align: simpletable.AlignCenter, Text: isAllowedToRemove(item.resourceType)})
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
