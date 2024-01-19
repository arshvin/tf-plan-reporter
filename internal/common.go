package internal

import (
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"

	tfJson "github.com/hashicorp/terraform-json"
)

type ResourceData struct {
	Type  string
	Name  string
	Index string
}

type ConsolidatedJson struct {
	Created   []*ResourceData
	Updated   []*ResourceData
	Deleted   []*ResourceData
	Unchanged []*ResourceData
}

// totalItems function returns amount of all items in the consolidatedJson struct
func (cj *ConsolidatedJson) TotalItems() int {
	return len(cj.Created) + len(cj.Updated) + len(cj.Deleted) + len(cj.Unchanged)
}

// parse function parses input data as parameter and puts it to consolidatedJson struct
func (cj *ConsolidatedJson) Parse(entity *tfJson.Plan) {
	for _, resource := range entity.ResourceChanges {

		var resourceIndex string
		switch resource.Index.(type) {
		case int:
			resourceIndex = fmt.Sprintf("%d", resource.Index)
		case string:
			resourceIndex = fmt.Sprintf("%s", resource.Index)
		case nil:
			resourceIndex = ""
		default:
			resourceIndex = fmt.Sprint(resource.Index)
		}

		resourceItem := &ResourceData{
			Type:  resource.Type,
			Name:  resource.Name,
			Index: resourceIndex,
		}

		tableRecordContext := log.WithFields(log.Fields{
			"resource_type":  resourceItem.Type,
			"resource_name":  resourceItem.Name,
			"resource_index": resourceItem.Index,
		})
		tableRecordContext.Debug("Created new resource item of report table")

		switch {
		case slices.Contains(resource.Change.Actions, tfJson.ActionCreate):
			cj.Created = append(cj.Created, resourceItem)

			tableRecordContext.Debug("The item has been put to 'Created' list")

		case slices.Contains(resource.Change.Actions, tfJson.ActionDelete):
			cj.Deleted = append(cj.Deleted, resourceItem)

			tableRecordContext.Debug("The item has been put to 'Deleted' list")

		case slices.Contains(resource.Change.Actions, tfJson.ActionUpdate):
			cj.Updated = append(cj.Updated, resourceItem)

			tableRecordContext.Debug("The item has been put to 'Updated' list")

		case slices.Contains(resource.Change.Actions, tfJson.ActionNoop):
			cj.Unchanged = append(cj.Unchanged, resourceItem)

			tableRecordContext.Debug("The item has been put to 'Unchanged' list")
		}
	}
}
