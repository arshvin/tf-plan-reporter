package processing

import (
	"fmt"

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

func formatMainContent(items []*resourceData) *simpletable.Table {
	headers := []string{"Type", "Index", "Name"}

	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{},
	}

	for _, header := range headers {
		table.Header.Cells = append(
			table.Header.Cells, &simpletable.Cell{Align: simpletable.AlignCenter, Text: header},
		)
	}

	for _, item := range items {
		row := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: item.resourceType},
			{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s", item.resourceIndex)},
			{Align: simpletable.AlignLeft, Text: item.resourceName},
		}

		table.Body.Cells = append(table.Body.Cells, row)
	}

	return table
}
