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
	User          User                         `belongs_to:"user" json:"-"`
	UserID        uuid.UUID                    `json:"-" db:"user_id"`
	Participant   Participant                  `belongs_to:"participant" json:"-"`
	ParticipantID uuid.UUID                    `json:"-" db:"participant_id"`
	Screening     Screening                    `belongs_to:"screening" json:"-"`
	ScreeningID   uuid.UUID                    `json:"-" db:"screening_id"`
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
	r.ReferralData.DateOfAttendance.CalculatedDate = r.ReferralData.DateOfAttendance.GivenDate
	if r.ReferralData.DateOfAttendance.Calendar == "thai" {
		r.ReferralData.DateOfAttendance.CalculatedDate = r.ReferralData.DateOfAttendance.CalculatedDate.AddDate(-543, 0, 0)
	}
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
