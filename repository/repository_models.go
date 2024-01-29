package repository

import (
	"time"
)

type IDatabaseObject interface {
	GetUpdatedAt() time.Time
	SetUpdatedAt(time time.Time)
	GetCreatedAt() time.Time
	SetCreatedAt(time time.Time)
	GetIsActive() bool
	SetIsActive(isActive bool)
}

type DatabaseObject struct {
	UpdatedAt time.Time
	CreatedAt time.Time
	IsActive  bool
}

func (do *DatabaseObject) GetUpdatedAt() time.Time {
	return do.UpdatedAt
}

func (do *DatabaseObject) SetUpdatedAt(time time.Time) {
	do.UpdatedAt = time
}

func (do *DatabaseObject) GetCreatedAt() time.Time {
	return do.CreatedAt
}

func (do *DatabaseObject) SetCreatedAt(time time.Time) {
	do.CreatedAt = time
}

func (do *DatabaseObject) GetIsActive() bool {
	return do.IsActive
}

func (do *DatabaseObject) SetIsActive(isActive bool) {
	do.IsActive = isActive
}
