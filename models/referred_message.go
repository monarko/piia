package models

import (
	"encoding/json"
	"time"

	"github.com/monarko/piia/helpers/types"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
)

// ReferredMessage model
type ReferredMessage struct {
	ID            uuid.UUID                    `json:"id" db:"id"`
	CreatedAt     time.Time                    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time                    `json:"updated_at" db:"updated_at"`
	ParticipantID uuid.UUID                    `json:"participant_id" db:"participant_id"`
	ScreeningID   uuid.UUID                    `json:"screening_id" db:"screening_id"`
	UserID        uuid.UUID                    `json:"user_id" db:"user_id"`
	Message       string                       `json:"message" db:"message"`
	ReferralData  types.ReferredMessageElement `json:"referral_data" db:"referral_data"`
}

// String is not required by pop and may be deleted
func (r ReferredMessage) String() string {
	jr, _ := json.Marshal(r)
	return string(jr)
}

// ReferredMessages is not required by pop and may be deleted
type ReferredMessages []ReferredMessage

// String is not required by pop and may be deleted
func (r ReferredMessages) String() string {
	jr, _ := json.Marshal(r)
	return string(jr)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (r *ReferredMessage) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (r *ReferredMessage) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (r *ReferredMessage) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
