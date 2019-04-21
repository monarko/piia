package models

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
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
	Screener       User                          `belongs_to:"user" json:"-"`
	ScreenerID     uuid.UUID                     `json:"-" db:"screener_id"`
	Participant    Participant                   `belongs_to:"participant" json:"-"`
	ParticipantID  uuid.UUID                     `json:"-" db:"participant_id"`
	Notifications  Notifications                 `has_many:"notifications" json:"-"`
	OverReadings   OverReadings                  `has_many:"over_readings" json:"-"`
}

// String is not required by pop and may be deleted
func (s Screening) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// Maps will return a map
func (s Screening) Maps() map[string]interface{} {
	bt, _ := json.Marshal(s)
	m := make(map[string]interface{})
	json.Unmarshal(bt, &m)
	delete(m, "screener")
	delete(m, "participant")
	delete(m, "created_at")
	delete(m, "updated_at")
	return m
}

// Screenings is not required by pop and may be deleted
type Screenings []Screening

// String is not required by pop and may be deleted
func (s Screenings) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// SectionStatus object
type SectionStatus struct {
	Section string `json:"section"`
	Done    bool   `json:"done"`
}

// Status object
type Status struct {
	Sections  []SectionStatus `json:"sections"`
	Completed bool            `json:"completed"`
}

// Statuses returns status for all parts
func (s Screening) Statuses() Status {
	diabetes := SectionStatus{"Diabetes", false}
	medicalHistory := SectionStatus{"Medical History", false}
	medications := SectionStatus{"Medications", false}
	measurements := SectionStatus{"Measurements", false}
	pathology := SectionStatus{"Pathology", false}
	eyeAssessments := SectionStatus{"Eye Assessment", false}

	completed := false

	if len(s.Diabetes.DiabetesType.String) > 0 && s.Diabetes.Duration.Valid {
		diabetes.Done = true
	}

	if s.Medications.TakingMedications.Valid && s.Medications.OnInsulin.Valid {
		medications.Done = true
	}

	if s.Measurements.BloodPressure.SBP.Valid && s.Measurements.BloodPressure.DBP.Valid {
		measurements.Done = true
	}

	if s.Pathology.HbA1C.HbA1C.Valid && s.Pathology.Lipids.TotalCholesterol.Valid {
		pathology.Done = true
	}

	if len(s.Eyes.RightEye.VisualAcuity.String) > 0 && len(s.Eyes.RightEye.DRGrading.String) > 0 && len(s.Eyes.RightEye.DMEAssessment.String) > 0 && len(s.Eyes.LeftEye.VisualAcuity.String) > 0 && len(s.Eyes.LeftEye.DRGrading.String) > 0 && len(s.Eyes.LeftEye.DMEAssessment.String) > 0 {
		eyeAssessments.Done = true
	}

	if s.MedicalHistory.Smoker.Valid && (s.MedicalHistory.Morbidities != nil || (diabetes.Done && medications.Done && measurements.Done && pathology.Done && eyeAssessments.Done)) {
		medicalHistory.Done = true
	}

	if diabetes.Done && medicalHistory.Done && medications.Done && measurements.Done && pathology.Done && eyeAssessments.Done {
		completed = true
	}

	statuses := []SectionStatus{diabetes, medicalHistory, medications, measurements, pathology, eyeAssessments}
	all := Status{statuses, completed}

	return all
}

// StatusesMap returns status for all parts in Map
func (s Screening) StatusesMap() map[string]bool {
	all := s.Statuses()

	maps := make(map[string]bool)

	for _, s := range all.Sections {
		maps[s.Section] = s.Done
	}

	return maps
}

// DaysAgo returns days ago its updated
func (s Screening) DaysAgo() int {
	days := int(math.Floor(time.Now().Sub(s.UpdatedAt).Hours() / 24))

	return days
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (s *Screening) Validate(tx *pop.Connection) (*validate.Errors, error) {
	valids := make([]validate.Validator, 0)

	s.Measurements.BloodPressure.AssessmentDate.CalculatedDate = s.Measurements.BloodPressure.AssessmentDate.GivenDate
	if s.Measurements.BloodPressure.AssessmentDate.Calendar == "thai" {
		s.Measurements.BloodPressure.AssessmentDate.CalculatedDate = s.Measurements.BloodPressure.AssessmentDate.CalculatedDate.AddDate(-543, 0, 0)
	}

	s.Pathology.HbA1C.AssessmentDate.CalculatedDate = s.Pathology.HbA1C.AssessmentDate.GivenDate
	if s.Pathology.HbA1C.AssessmentDate.Calendar == "thai" {
		s.Pathology.HbA1C.AssessmentDate.CalculatedDate = s.Pathology.HbA1C.AssessmentDate.CalculatedDate.AddDate(-543, 0, 0)
	}

	s.Pathology.Lipids.AssessmentDate.CalculatedDate = s.Pathology.Lipids.AssessmentDate.GivenDate
	if s.Pathology.Lipids.AssessmentDate.Calendar == "thai" {
		s.Pathology.Lipids.AssessmentDate.CalculatedDate = s.Pathology.Lipids.AssessmentDate.CalculatedDate.AddDate(-543, 0, 0)
	}

	s.Eyes.AssessmentDate.CalculatedDate = s.Eyes.AssessmentDate.GivenDate
	if s.Eyes.AssessmentDate.Calendar == "thai" {
		s.Eyes.AssessmentDate.CalculatedDate = s.Eyes.AssessmentDate.CalculatedDate.AddDate(-543, 0, 0)
	}

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

// IsValid checks if is valid or not
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

// IsValid checks if is valid or not
func (v *InRangeInt) IsValid(errors *validate.Errors) {
	if v.Field.Valid {
		if v.Field.Int < v.Start || v.Field.Int > v.End {
			errors.Add(validators.GenerateKey(v.Name), fmt.Sprintf("%s acceptable range is between %d - %d", v.Name, v.Start, v.End))
		}
	}
}
