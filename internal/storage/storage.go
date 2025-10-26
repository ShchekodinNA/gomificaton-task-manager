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
}

func getDefaultStoragePath() (string, error) {
	configDir, err := utils.EnsureAppDataLocation()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}
	configPath := filepath.Join(configDir, "data.db")

	return configPath, nil
}

func NewSqlliteStorage() (*Storage, error) {
	storagePath, err := getDefaultStoragePath()
	if err != nil {
		return nil, fmt.Errorf("get default storage path: %w", err)
	}

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?cache=shared&mode=rwc&_fk=1", storagePath))

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Storage{db: db}, nil
}
