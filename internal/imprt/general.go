package imprt

import "gomificator/internal/models"

type Importer interface {
	Import() ([]models.TimerModel, error)
}
