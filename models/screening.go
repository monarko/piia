package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/nulls"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
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
	Notifications  Notifications                 `has_many:"notifications" json:"-"`
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
	valids := make([]validate.Validator, 0)

	valids = append(valids, &InRangeFloat64{Field: s.Pathology.HbA1C.HbA1C, Name: "HbA1C", Start: 3.0, End: 30.0})
	valids = append(valids, &InRangeInt{Field: s.Measurements.BloodPressure.SBP, Name: "SBP", Start: 80, End: 250})
	valids = append(valids, &InRangeInt{Field: s.Measurements.BloodPressure.DBP, Name: "DBP", Start: 50, End: 180})

	return validate.Validate(valids...), nil
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

// InRangeFloat64 Check
type InRangeFloat64 struct {
	Name  string
	Field nulls.Float64
	Start float64
	End   float64
}

// IsValid checks if username is valid or not
func (v *InRangeFloat64) IsValid(errors *validate.Errors) {
	if v.Field.Valid {
		if v.Field.Float64 < v.Start || v.Field.Float64 > v.End {
			errors.Add(validators.GenerateKey(v.Name), fmt.Sprintf("%s acceptable range is between %.1f - %.1f", v.Name, v.Start, v.End))
		}
	}
}

// InRangeInt Check
type InRangeInt struct {
	Name  string
	Field nulls.Int
	Start int
	End   int
}

// IsValid checks if username is valid or not
func (v *InRangeInt) IsValid(errors *validate.Errors) {
	if v.Field.Valid {
		if v.Field.Int < v.Start || v.Field.Int > v.End {
			errors.Add(validators.GenerateKey(v.Name), fmt.Sprintf("%s acceptable range is between %d - %d", v.Name, v.Start, v.End))
		}
	}
}
