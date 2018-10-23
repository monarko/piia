package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
)

// Notification model
type Notification struct {
	ID            uuid.UUID   `json:"id" db:"id"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`
	Type          string      `json:"type" db:"type"`
	Message       string      `json:"message" db:"message"`
	FromUser      User        `belongs_to:"user" json:"from_user"`
	FromUserID    uuid.UUID   `json:"-" db:"from_user_id"`
	Participant   Participant `belongs_to:"participant" json:"participant"`
	ParticipantID uuid.UUID   `json:"-" db:"participant_id"`
	Screening     Screening   `belongs_to:"screening" json:"screening"`
	ScreeningID   uuid.UUID   `json:"-" db:"screening_id"`
	Status        string      `json:"status" db:"status"`
	Site          string      `json:"site" db:"site"`
}

// String is not required by pop and may be deleted
func (n Notification) String() string {
	jn, _ := json.Marshal(n)
	return string(jn)
}

// Notifications is not required by pop and may be deleted
type Notifications []Notification

// String is not required by pop and may be deleted
func (n Notifications) String() string {
	jn, _ := json.Marshal(n)
	return string(jn)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (n *Notification) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (n *Notification) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (n *Notification) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// NotClosedNotifications returns the notifications without "closed" status
func NotClosedNotifications() pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		return q.Where("status != ?", "closed")
	}
}
