package storage

import (
	"database/sql"
	"fmt"
	"gomificator/internal/constnats"
	"gomificator/internal/models"
	"time"
)

type TimerRepository interface {
	Save(timer models.TimerModel) (int, error) // Создает новый, если id == nil или обновляет нужную запись
	GetLastTimers(q int) ([]models.TimerModel, error)
	GetTimersBetweenDates(startDate, endDate time.Time) ([]models.TimerModel, error)
	Delete(id int) error
}

type timerRepository struct {
	db *sql.DB
}

func NewTimerRepository(db *sql.DB) TimerRepository {
	return &timerRepository{db: db}
}

func (r *timerRepository) Save(timer models.TimerModel) (int, error) {
	if timer.Id != nil {
		return r.update(timer)
	}
	if timer.Id == nil && timer.ExternalId != nil {
		id, err := r.getIdByExternalId(*timer.ExternalId)
		if err != nil {
			return r.create(timer)
		}
		timer.Id = &id
		return r.update(timer)
	}
	return r.create(timer)
}

func (r *timerRepository) create(t models.TimerModel) (int, error) {

	res, err := r.db.Exec(`
		INSERT INTO timers (external_id, fixed_at, seconds_spent, name, description)
		VALUES (?, ?, ?, ?, ?)`,
		*t.ExternalId,
		t.FixatedAt.Format("2006-01-02"),
		int(t.SecondsSpent.Seconds()),
		t.Name,
		t.Description,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	return int(id), err
}

func (r *timerRepository) update(t models.TimerModel) (int, error) {
	_, err := r.db.Exec(`
		UPDATE timers
		SET external_id = ?,
			fixed_at = ?,
			seconds_spent = ?,
			name = ?,
			description = ?
		WHERE id = ?`,
		*t.ExternalId,
		t.FixatedAt.Format("2006-01-02"),
		int(t.SecondsSpent.Seconds()),
		t.Name,
		t.Description,
		*t.Id,
	)
	return *t.Id, err
}

func (r *timerRepository) getIdByExternalId(externalId string) (int, error) {
	query := "SELECT id FROM timers WHERE external_id = ? limit 1"

	var id int
	err := r.db.QueryRow(
		query,
		externalId,
	).Scan(&id)
	return id, err
}

func (r *timerRepository) GetLastTimers(q int) ([]models.TimerModel, error) {
	rows, err := r.db.Query(`
		SELECT id, external_id, fixed_at, seconds_spent, name, description, created_at
		FROM timers
		ORDER BY created_at DESC
		LIMIT ?`, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timers []models.TimerModel
	for rows.Next() {
		var t models.TimerModel
		var fixatedAtStr string
		var secondsSpent int
		var externalId sql.NullString
		var createdAt time.Time

		err := rows.Scan(
			&t.Id,
			&externalId,
			&fixatedAtStr,
			&secondsSpent,
			&t.Name,
			&t.Description,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		if externalId.Valid {
			t.ExternalId = &externalId.String
		}
		t.FixatedAt, _ = time.Parse(constnats.DateLayout, fixatedAtStr)
		t.SecondsSpent = time.Duration(secondsSpent) * time.Second
		t.CreatedAt = &createdAt

		timers = append(timers, t)
	}
	return timers, nil
}

func (r *timerRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM timers WHERE id = ?", id)
	return err
}

func (r *timerRepository) GetTimersBetweenDates(startDate, endDate time.Time) ([]models.TimerModel, error) {
	rows, err := r.db.Query(`
		SELECT id, external_id, fixed_at, seconds_spent, name, description, created_at
		FROM timers
		WHERE fixed_at BETWEEN ? AND ?`,
		startDate.Format(constnats.DateLayout),
		endDate.Format(constnats.DateLayout),
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var timers []models.TimerModel

	for rows.Next() {
		var t models.TimerModel
		var fixatedAtStr string
		var secondsSpent int
		var externalId sql.NullString
		var createdAt time.Time

		err := rows.Scan(
			&t.Id,
			&externalId,
			&fixatedAtStr,
			&secondsSpent,
			&t.Name,
			&t.Description,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("row scan: %w", err)
		}

		if externalId.Valid {
			t.ExternalId = &externalId.String
		}

		t.CreatedAt = &createdAt
		t.FixatedAt, _ = time.Parse(constnats.DateLayout, fixatedAtStr)
		t.SecondsSpent = time.Duration(secondsSpent) * time.Second

		timers = append(timers, t)

	}

	return timers, nil
}
