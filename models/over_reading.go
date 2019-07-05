package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/monarko/piia/helpers/types"
)

// OverReading model
type OverReading struct {
	ID            uuid.UUID               `json:"id" db:"id"`
	CreatedAt     time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at" db:"updated_at"`
	Eyes          types.EyeOverReading    `json:"eyes" db:"eye_assessment"`
	Referral      types.ReferralScreening `json:"referral" db:"referral"`
	OverReader    User                    `belongs_to:"user" json:"-"`
	OverReaderID  uuid.UUID               `json:"-" db:"over_reader_id"`
	Participant   Participant             `belongs_to:"participant" json:"-"`
	ParticipantID uuid.UUID               `json:"-" db:"participant_id"`
	Screening     Screening               `belongs_to:"screening" json:"-"`
	ScreeningID   uuid.UUID               `json:"-" db:"screening_id"`
}

// String is not required by pop and may be deleted
func (o OverReading) String() string {
	jo, _ := json.Marshal(o)
	return string(jo)
}

// Maps will return a map
func (o OverReading) Maps() map[string]interface{} {
	bt, _ := json.Marshal(o)
	m := make(map[string]interface{})
	json.Unmarshal(bt, &m)
	delete(m, "screening")
	delete(m, "participant")
	delete(m, "over_reader")
	delete(m, "created_at")
	delete(m, "updated_at")
	return m
}

// OverReadings is not required by pop and may be deleted
type OverReadings []OverReading

// String is not required by pop and may be deleted
func (o OverReadings) String() string {
	jo, _ := json.Marshal(o)
	return string(jo)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (o *OverReading) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: o.Eyes.LeftEye.DRGrading.String, Name: "DR Grade for Left Eye"},
		&validators.StringIsPresent{Field: o.Eyes.LeftEye.DMEAssessment.String, Name: "DME Assessment for Left Eye"},
		&validators.StringIsPresent{Field: o.Eyes.RightEye.DRGrading.String, Name: "DR Grade for Right Eye"},
		&validators.StringIsPresent{Field: o.Eyes.RightEye.DMEAssessment.String, Name: "DME Assessment for Right Eye"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (o *OverReading) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (o *OverReading) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
