package storage

import (
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func MigrateDb(strg *Storage) error {
	goose.SetBaseFS(migrationFS)
	if err := goose.Up(strg.db, "migrations"); err != nil {
		return fmt.Errorf("migrate db: %w", err)
	}

	return nil
}
