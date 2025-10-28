package imprt

import "fmt"

type ImporterType string

const (
	ImporterTypeSuperProductivityExport ImporterType = "spexport"
	ImporterTypeSuperProductivityBackup ImporterType = "spbackup"
)

func GetImporterTypeByString(importerTypeStr string) (ImporterType, error) {
	switch importerTypeStr {
	case string(ImporterTypeSuperProductivityExport):
		return ImporterTypeSuperProductivityExport, nil
	case string(ImporterTypeSuperProductivityBackup):
		return ImporterTypeSuperProductivityBackup, nil
	default:
		return "", fmt.Errorf("unknown importer type: %s", importerTypeStr)
	}
}
