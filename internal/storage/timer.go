package storage

import (
	"database/sql"
	"time"
)

type timerModel struct {
	Id          *int
	ExternalId  *string
	Name        string
	Description string
	CreatedAt   time.Time
	FixatedAt   time.Time
	SecondSpent time.Duration
}

type TimerRepository interface {
	Save(timer timerModel) (int, error) // Создает новый, если id == nil или обновляет нужную запись
	GetLastTimers(q int) ([]timerModel, error)
	Delete(id int) error
}

type timerRepository struct {
	db *sql.DB
}

func NewTimerRepository(db *sql.DB) TimerRepository {
	return &timerRepository{db: db}
}

func (r *timerRepository) Save(timer timerModel) (int, error) {
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

func (r *timerRepository) create(t timerModel) (int, error) {

	res, err := r.db.Exec(`
		INSERT INTO timers (external_id, fixed_at, seconds_spent, name, description)
		VALUES (?, ?, ?, ?, ?)`,
		*t.ExternalId,
		t.FixatedAt.Format("2006-01-02"),
		int(t.SecondSpent.Seconds()),
		t.Name,
		t.Description,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	return int(id), err
}

func (r *timerRepository) update(t timerModel) (int, error) {
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
		int(t.SecondSpent.Seconds()),
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

func (r *timerRepository) GetLastTimers(q int) ([]timerModel, error) {
	rows, err := r.db.Query(`
		SELECT id, external_id, fixed_at, seconds_spent, name, description, created_at
		FROM timers
		ORDER BY created_at DESC
		LIMIT ?`, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timers []timerModel
	for rows.Next() {
		var t timerModel
		var fixatedAtStr string
		var secondsSpent int
		var externalId sql.NullString

		err := rows.Scan(
			&t.Id,
			&externalId,
			&fixatedAtStr,
			&secondsSpent,
			&t.Name,
			&t.Description,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if externalId.Valid {
			t.ExternalId = &externalId.String
		}
		t.FixatedAt, _ = time.Parse("2006-01-02", fixatedAtStr)
		t.SecondSpent = time.Duration(secondsSpent) * time.Second

		timers = append(timers, t)
	}
	return timers, nil
}

func (r *timerRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM timers WHERE id = ?", id)
	return err
}
