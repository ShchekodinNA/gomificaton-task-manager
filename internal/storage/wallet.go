package storage

import (
	"database/sql"
	"fmt"
	"gomificator/internal/constnats"
	"gomificator/internal/models"
)

type WalletRepository interface {
	Load() (models.WalletModel, error)
	Save(wallet models.WalletModel) error
}

type walletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Load() (models.WalletModel, error) {
	rows, err := r.db.Query("SELECT medal_type, count FROM wallet")
	if err != nil {
		return nil, fmt.Errorf("query wallet: %w", err)
	}
	defer rows.Close()

	res := make(models.WalletModel)

	for rows.Next() {
		var medalStr string
		var cnt int
		if err := rows.Scan(&medalStr, &cnt); err != nil {
			return nil, fmt.Errorf("scan wallet row: %w", err)
		}
		medal, err := constnats.LoadMedal(medalStr)
		if err != nil {
			return nil, fmt.Errorf("load medal: %w", err)
		}
		res[medal] = cnt
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return res, nil
}

func (r *walletRepository) Save(wallet models.WalletModel) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		// Rollback if still active; ignore error if already committed
		_ = tx.Rollback()
	}()

	stmt := `
        INSERT INTO wallet (medal_type, count)
        VALUES (?, ?)
        ON CONFLICT(medal_type) DO UPDATE SET count = excluded.count`

	for medal, cnt := range wallet {
		if _, err := tx.Exec(stmt, string(medal), cnt); err != nil {
			return fmt.Errorf("upsert medal %s: %w", medal, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
