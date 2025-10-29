package imprt

import (
	"encoding/json"
	"fmt"
	"gomificator/internal/constnats"
	"gomificator/internal/models"
	"os"
	"time"
)

const superProductivity = "SuperProductivity"

type backupFileContent struct {
	Data dataContent `json:"data"`
}

type dataContent struct {
	Task         taskEntry    `json:"task"`
	ArchiveOld   archiveEntry `json:"archiveOld"`
	ArchiveYoung archiveEntry `json:"archiveYoung"`
}

type archiveEntry struct {
	Task taskEntry `json:"task"`
}

type taskEntry struct {
	Ids      []string              `json:"ids"`
	Entities map[string]taskEntity `json:"entities"`
}

type taskEntity struct {
	TimeSpentOnDay map[string]int `json:"timeSpentOnDay"`
	Title          string         `json:"title"`
	Description    string         `json:"notes"`
}

type impoerterSuperProductivityExportFile struct {
	exportFile *os.File
}

func NewImporterFromSuperProductivityExportFile(file *os.File) Importer {
	return &impoerterSuperProductivityExportFile{
		exportFile: file,
	}
}

func (i *impoerterSuperProductivityExportFile) Import() ([]models.TimerModel, error) {

	var backupContent backupFileContent
	decoder := json.NewDecoder(i.exportFile)

	if err := decoder.Decode(&backupContent); err != nil {
		return nil, fmt.Errorf("decode backup file: %w", err)
	}

	allEntities := extractAllEntities(backupContent.Data)

	timers, err := createBatchTimers(allEntities)
	if err != nil {
		return nil, fmt.Errorf("create batch timers: %w", err)
	}

	return timers, nil

}

type impoerterSuperProductivityBackupFile struct {
	backupFile *os.File
}

func NewImporterFromSuperProductivityBackupFile(file *os.File) Importer {
	return &impoerterSuperProductivityBackupFile{
		backupFile: file,
	}
}

func (i *impoerterSuperProductivityBackupFile) Import() ([]models.TimerModel, error) {
	var dataContent dataContent

	decoder := json.NewDecoder(i.backupFile)

	if err := decoder.Decode(&dataContent); err != nil {
		return nil, fmt.Errorf("decode backup file: %w", err)
	}

	allEntities := extractAllEntities(dataContent)

	timers, err := createBatchTimers(allEntities)
	if err != nil {
		return nil, fmt.Errorf("create batch timers: %w", err)
	}

	return timers, nil

}

// general functions
func extractAllEntities(dc dataContent) map[string]taskEntity {
	total := 0
	if dc.Task.Entities != nil {
		total += len(dc.Task.Entities)
	}
	if dc.ArchiveOld.Task.Entities != nil {
		total += len(dc.ArchiveOld.Task.Entities)
	}
	if dc.ArchiveYoung.Task.Entities != nil {
		total += len(dc.ArchiveYoung.Task.Entities)
	}

	out := make(map[string]taskEntity, total)

	for k, v := range dc.Task.Entities {
		out[k] = v
	}
	for k, v := range dc.ArchiveOld.Task.Entities {
		out[k] = v
	}
	for k, v := range dc.ArchiveYoung.Task.Entities {
		out[k] = v
	}

	return out
}

func createBatchTimers(entities map[string]taskEntity) ([]models.TimerModel, error) {
	var timers []models.TimerModel

	for taskId, task := range entities {
		for dateStr, seconds := range task.TimeSpentOnDay {
			var externalIdPtr string = generateExternalId(taskId, dateStr)
			fixatedAt, err := time.Parse(constnats.DateLayout, dateStr)
			if err != nil {
				return nil, fmt.Errorf("parse date %s: %w", dateStr, err)
			}
			timers = append(timers, models.TimerModel{
				Id:           nil,
				ExternalId:   &externalIdPtr,
				CreatedAt:    nil,
				Name:         task.Title,
				Description:  task.Description,
				SecondsSpent: time.Millisecond * time.Duration(seconds),
				FixatedAt:    fixatedAt,
			})
		}
	}

	return timers, nil
}

func generateExternalId(taskId string, dateStr string) string {
	return fmt.Sprintf("%s:%s:%s", superProductivity, taskId, dateStr)
}
