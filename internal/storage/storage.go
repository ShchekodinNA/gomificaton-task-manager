package storage

import (
	"database/sql"
	"fmt"
	"gomificator/internal/utils"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB

	Timers TimerRepository
}

// NewSqlliteStorage creates a new SQLite storage instance.
// It initializes a SQLite database connection using the default storage path.
//
// Returns:
//   - *Storage: pointer to the initialized Storage instance
//   - error: any error encountered during initialization, including:
//   - failure to determine default storage path
//   - failure to open database connection
func NewSqlliteStorage() (*Storage, error) {
	storagePath, err := getDefaultStoragePath()
	if err != nil {
		return nil, fmt.Errorf("get default storage path: %w", err)
	}

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?cache=shared&mode=rwc&_fk=1", storagePath))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	timerRepo := NewTimerRepository(db)

	return &Storage{db: db, Timers: timerRepo}, nil
}

func getDefaultStoragePath() (string, error) {
	configDir, err := utils.EnsureAppDataLocation()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}
	configPath := filepath.Join(configDir, "data.db")

	return configPath, nil
}

// IsStorageExists checks if the default storage location exists.
//
// Returns:
//   - bool: true if storage exists, false otherwise
//   - error: potential errors that may occur during the check:
//   - error getting default storage path
//   - error checking file existence
func IsStorageExists() (bool, error) {
	location, err := getDefaultStoragePath()
	if err != nil {
		return false, fmt.Errorf("get app data location: %w", err)
	}

	isExists, err := utils.FileExists(location)
	if err != nil {
		return false, fmt.Errorf("file exists: %w", err)
	}

	return isExists, nil

}
