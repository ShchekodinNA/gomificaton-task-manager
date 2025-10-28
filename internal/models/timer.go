package models

import "time"

type TimerModel struct {
	Id           *int
	ExternalId   *string
	CreatedAt    *time.Time
	Name         string
	Description  string
	FixatedAt    time.Time
	SecondsSpent time.Duration
}
