package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/monarko/piia/helpers/types"
)

// Screening model
type Screening struct {
	ID             uuid.UUID                     `json:"id" db:"id"`
	CreatedAt      time.Time                     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time                     `json:"updated_at" db:"updated_at"`
	Diabetes       types.DiabetesScreening       `json:"diabetes" db:"diabetes"`
	MedicalHistory types.MedicalHistoryScreening `json:"medical_history" db:"medical_history"`
	Medications    types.MedicationScreening     `json:"medications" db:"medications"`
	Measurements   types.MeasurementScreening    `json:"measurements" db:"measurements"`
	Pathology      types.PathologyScreening      `json:"pathology" db:"pathology"`
	Eyes           types.EyeScreening            `json:"eyes" db:"eye"`
	Referral       types.ReferralScreening       `json:"referral" db:"referral"`
	Screener       User                          `belongs_to:"user" json:"screener"`
	ScreenerID     uuid.UUID                     `json:"-" db:"screener_id"`
	Participant    Participant                   `belongs_to:"participant" json:"participant"`
	ParticipantID  uuid.UUID                     `json:"-" db:"participant_id"`
}

// String is not required by pop and may be deleted
func (s Screening) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// Screenings is not required by pop and may be deleted
type Screenings []Screening

// String is not required by pop and may be deleted
func (s Screenings) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (s *Screening) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
	// &validators.StringIsPresent{Field: s.Eyes.LeftEye.VisualAcuity, Name: "LeftVisualAcuity"},
	// &validators.StringIsPresent{Field: s.Eyes.LeftEye.DRGrading, Name: "LeftGradingDr"},
	// &validators.StringIsPresent{Field: s.Eyes.LeftEye.DMEAssessment, Name: "LeftGradingDme"},
	// &validators.StringIsPresent{Field: s.Eyes.RightEye.VisualAcuity, Name: "RightVisualAcuity"},
	// &validators.StringIsPresent{Field: s.Eyes.RightEye.DRGrading, Name: "RightGradingDr"},
	// &validators.StringIsPresent{Field: s.Eyes.RightEye.DMEAssessment, Name: "RightGradingDme"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (s *Screening) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (s *Screening) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
