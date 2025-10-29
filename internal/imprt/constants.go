package imprt

import "fmt"

type ImporterType int

const (
	ImporterTypeSuperProductivityExport ImporterType = iota
	ImporterTypeSuperProductivityBackup
)

var importerTypeMap = map[ImporterType]string{
	ImporterTypeSuperProductivityExport: "spexport",
	ImporterTypeSuperProductivityBackup: "spbackup",
}

func (i ImporterType) String() string {
	return importerTypeMap[i]
}

// ImporterTypeSuperProductivityExport ImporterType = "spexport"
// ImporterTypeSuperProductivityBackup ImporterType = "spbackup"

func GetImporterTypeByString(importerTypeStr string) (ImporterType, error) {
	switch importerTypeStr {
	case importerTypeMap[ImporterTypeSuperProductivityExport]:
		return ImporterTypeSuperProductivityExport, nil
	case importerTypeMap[ImporterTypeSuperProductivityBackup]:
		return ImporterTypeSuperProductivityBackup, nil
	default:
		return 0, fmt.Errorf("unknown importer type: %s", importerTypeStr)
	}
}
