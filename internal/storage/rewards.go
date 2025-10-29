package storage

import (
    "database/sql"
    "fmt"
    "gomificator/internal/constnats"
    "gomificator/internal/models"
    "time"
)

type RewardsDailyRepository interface {
    LoadByDate(day time.Time) (models.WalletModel, error)
    ReplaceForDate(day time.Time, daily models.WalletModel) error
}

type rewardsDailyRepository struct {
    db *sql.DB
}

func NewRewardsDailyRepository(db *sql.DB) RewardsDailyRepository {
    return &rewardsDailyRepository{db: db}
}

func (r *rewardsDailyRepository) LoadByDate(day time.Time) (models.WalletModel, error) {
    rows, err := r.db.Query(`SELECT medal_type, count FROM rewards_daily WHERE day = ?`, day.Format(constnats.DateLayout))
    if err != nil {
        return nil, fmt.Errorf("query rewards_daily: %w", err)
    }
    defer rows.Close()

    out := make(models.WalletModel)
    for rows.Next() {
        var medal string
        var cnt int
        if err := rows.Scan(&medal, &cnt); err != nil {
            return nil, fmt.Errorf("scan rewards_daily: %w", err)
        }
        m, err := constnats.LoadMedal(medal)
        if err != nil {
            return nil, fmt.Errorf("load medal: %w", err)
        }
        out[m] = cnt
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate rewards_daily: %w", err)
    }
    return out, nil
}

func (r *rewardsDailyRepository) ReplaceForDate(day time.Time, daily models.WalletModel) error {
    tx, err := r.db.Begin()
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer func() { _ = tx.Rollback() }()

    if _, err := tx.Exec(`DELETE FROM rewards_daily WHERE day = ?`, day.Format(constnats.DateLayout)); err != nil {
        return fmt.Errorf("delete old rewards_daily: %w", err)
    }

    if len(daily) > 0 {
        stmt := `INSERT INTO rewards_daily(day, medal_type, count) VALUES(?, ?, ?)`
        for medal, cnt := range daily {
            if cnt == 0 {
                continue
            }
            if _, err := tx.Exec(stmt, day.Format(constnats.DateLayout), string(medal), cnt); err != nil {
                return fmt.Errorf("insert rewards_daily %s: %w", medal, err)
            }
        }
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit tx: %w", err)
    }
    return nil
}

