package entity

import (
	"time"

	"github.com/google/uuid"
)

type Avatar struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ObjectKey string
	CreatedAt time.Time
	DeletedAt *time.Time
	URL       string
}
