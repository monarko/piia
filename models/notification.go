package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gofrs/uuid"
	"github.com/monarko/piia/helpers/types"
)

// Notification model
type Notification struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
	Type          string           `json:"type" db:"type"`
	Message       string           `json:"message" db:"message"`
	FromUser      User             `belongs_to:"user" json:"-"`
	FromUserID    uuid.UUID        `json:"from_user_id" db:"from_user_id"`
	Participant   Participant      `belongs_to:"participant" json:"-"`
	ParticipantID uuid.UUID        `json:"participant_id" db:"participant_id"`
	Screening     Screening        `belongs_to:"screening" json:"-"`
	ScreeningID   uuid.UUID        `json:"screening_id" db:"screening_id"`
	Status        string           `json:"status" db:"status"`
	Site          string           `json:"site" db:"site"`
	EventDate     types.CustomDate `json:"assessment_date" db:"event_date"`
}

// Maps will return a map
func (n Notification) Maps() map[string]interface{} {
	bt, _ := json.Marshal(n)
	m := make(map[string]interface{})
	json.Unmarshal(bt, &m)
	delete(m, "screening")
	delete(m, "participant")
	delete(m, "created_at")
	delete(m, "updated_at")
	return m
}

// Notifications is not required by pop and may be deleted
type Notifications []Notification

// NotClosedNotifications returns the notifications without "closed" status
func NotClosedNotifications() pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		return q.Where("status != ?", "closed")
	}
}
