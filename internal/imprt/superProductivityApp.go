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

	timers, err := createBatchTimers(backupContent.Data.Task.Entities)
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

	timers, err := createBatchTimers(dataContent.Task.Entities)
	if err != nil {
		return nil, fmt.Errorf("create batch timers: %w", err)
	}

	return timers, nil

}

// general functions
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
