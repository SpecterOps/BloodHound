package services

import (
	"database/sql"
	"errors"
	"time"
)

type Permission struct {
	Authority string       `json:"authority"`
	Name      string       `json:"name"`
	ID        int32        `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

// ErrNoPermissionFound indicates that there is no analysis request currently pending.
var ErrNoPermissionFound = errors.New("no permission was found")
