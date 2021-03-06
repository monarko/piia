package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
	"github.com/monarko/piia/helpers/types"
)

// ReferredMessage model
type ReferredMessage struct {
	ID            uuid.UUID                    `json:"id" db:"id"`
	CreatedAt     time.Time                    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time                    `json:"updated_at" db:"updated_at"`
	User          User                         `belongs_to:"user" json:"-"`
	UserID        uuid.UUID                    `json:"user_id" db:"user_id"`
	Participant   Participant                  `belongs_to:"participant" json:"-"`
	ParticipantID uuid.UUID                    `json:"participant_id" db:"participant_id"`
	Screening     Screening                    `belongs_to:"screening" json:"-"`
	ScreeningID   uuid.UUID                    `json:"screening_id" db:"screening_id"`
	Message       string                       `json:"message" db:"message"`
	ReferralData  types.ReferredMessageElement `json:"referral_data" db:"referral_data"`
}

// Maps will return a map
func (r ReferredMessage) Maps() map[string]interface{} {
	bt, _ := json.Marshal(r)
	m := make(map[string]interface{})
	json.Unmarshal(bt, &m)
	delete(m, "screening")
	delete(m, "participant")
	delete(m, "user")
	delete(m, "created_at")
	delete(m, "updated_at")
	return m
}

// ReferredMessages is not required by pop and may be deleted
type ReferredMessages []ReferredMessage

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (r *ReferredMessage) Validate(tx *pop.Connection) (*validate.Errors, error) {
	r.ReferralData.DateOfAttendance.CalculatedDate = r.ReferralData.DateOfAttendance.GivenDate
	if r.ReferralData.DateOfAttendance.Calendar == "thai" {
		r.ReferralData.DateOfAttendance.CalculatedDate = r.ReferralData.DateOfAttendance.CalculatedDate.AddDate(-543, 0, 0)
	}
	return validate.NewErrors(), nil
}
