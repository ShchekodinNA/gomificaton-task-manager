package storage

import (
	"database/sql"
	"errors"
	"time"
)

type timerModel struct {
	id    *int
	start time.Time
	end   time.Time
}

type TimerRepository interface {
	Save(timerModel) (int, error) // Создает новый, если id == nil или обновляет нужную запись
	GetLastTimers(q int) ([]timerModel, error)
	Delete(id int) error
}

type timerRepository struct {
	db *sql.DB
}

var (
	errTimerInvalid = errors.New("timer repository: invalid time range")
)

// NewTimerRepository provides a minimal constructor returning the interface.
func NewTimerRepository(db *sql.DB) TimerRepository {
	return &timerRepository{db: db}
}

// Save inserts or updates a timerModel and returns its id.
func (r *timerRepository) Save(m timerModel) (int, error) {
	if r == nil || r.db == nil {
		return 0, sql.ErrConnDone
	}

	if m.start.IsZero() || m.end.IsZero() || !m.end.After(m.start) {
		return 0, errTimerInvalid
	}

	start := m.start.UTC()
	end := m.end.UTC()

	if m.id == nil {
		res, err := r.db.Exec(`INSERT INTO timers (start_utc, end_utc) VALUES (?, ?)`, start, end)
		if err != nil {
			return 0, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		return int(id), nil
	}

	res, err := r.db.Exec(`UPDATE timers SET start_utc = ?, end_utc = ? WHERE id = ?`, start, end, *m.id)
	if err != nil {
		return 0, err
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return 0, sql.ErrNoRows
	}
	return *m.id, nil
}

// GetLastTimers returns the last q timers.
func (r *timerRepository) GetLastTimers(q int) ([]timerModel, error) {
	if r == nil || r.db == nil {
		return nil, sql.ErrConnDone
	}
	if q <= 0 {
		return []timerModel{}, nil
	}

	rows, err := r.db.Query(`SELECT id, start_utc, end_utc FROM timers ORDER BY start_utc DESC LIMIT ?`, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timers []timerModel
	for rows.Next() {
		var (
			id    int
			start time.Time
			end   time.Time
		)
		if err := rows.Scan(&id, &start, &end); err != nil {
			return nil, err
		}
		idCopy := id
		timers = append(timers, timerModel{
			id:    &idCopy,
			start: start,
			end:   end,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return timers, nil
}

// Delete removes a timer by id.
func (r *timerRepository) Delete(id int) error {
	if r == nil || r.db == nil {
		return sql.ErrConnDone
	}
	res, err := r.db.Exec(`DELETE FROM timers WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
