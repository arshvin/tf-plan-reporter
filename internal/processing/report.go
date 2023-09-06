package processing

type resourceData struct {
	resourceType string
	resourceName string
	resourceIndex string
}

type consolidatedJson struct {
	created []resourceData
	updated []resourceData
	deleted []resourceData
	unchanged []resourceData
}