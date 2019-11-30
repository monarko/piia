package models

import (
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/monarko/piia/helpers/types"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

// Participant object
type Participant struct {
	ID                  uuid.UUID        `json:"id" db:"id"`
	CreatedAt           time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at" db:"updated_at"`
	ParticipantID       string           `json:"participant_id" db:"participant_id"`
	Name                nulls.String     `json:"name" db:"name"`
	Gender              string           `json:"gender" db:"gender"`
	DOB                 types.CustomDate `json:"dob" db:"dob"`
	ContactNumber       nulls.String     `json:"contact_number" db:"contact_number"`
	IsEligible          bool             `json:"is_eligible" db:"is_eligible"`
	Consented           bool             `json:"consented" db:"consented"`
	IDType              nulls.String     `json:"id_type" db:"id_type"`
	IDNumber            nulls.String     `json:"id_number" db:"id_number"`
	User                User             `belongs_to:"user" json:"registrar"`
	UserID              uuid.UUID        `json:"author_id" db:"author_id"`
	Status              string           `json:"status" db:"status"`
	Screenings          Screenings       `has_many:"screenings" json:"-"`
	OverReadings        OverReadings     `has_many:"over_readings" json:"-"`
	Notifications       Notifications    `has_many:"notifications" json:"-"`
	Referrals           ReferredMessages `has_many:"referred_messages" json:"-"`
	ReferralAppointment bool             `json:"referral_appointment" db:"referral_appointment"`
}

// Maps will return a map
func (p Participant) Maps() map[string]interface{} {
	bt, _ := json.Marshal(p)
	m := make(map[string]interface{})
	json.Unmarshal(bt, &m)
	delete(m, "screenings")
	delete(m, "over_readings")
	delete(m, "notifications")
	delete(m, "referred_messages")
	delete(m, "created_at")
	delete(m, "updated_at")
	return m
}

// Participants holds the list of Participant
type Participants []Participant

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (p *Participant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	p.DOB.CalculatedDate = p.DOB.GivenDate
	if p.DOB.Calendar == "thai" && !p.DOB.GivenDate.IsZero() {
		p.DOB.CalculatedDate = p.DOB.CalculatedDate.AddDate(-543, 0, 0)
	}

	return validate.Validate(
		&validators.StringIsPresent{Field: p.Gender, Name: "Gender"},
		&validators.TimeIsPresent{Field: p.DOB.GivenDate, Name: "Date of birth"},
	), nil
}

// Completeness returns the completeness score
func (p Participant) Completeness() int {
	score, total := 0, 0

	// Participant
	if len(p.ParticipantID) == 9 {
		score += 10
	}
	gender := strings.ToLower(p.Gender)
	if gender == "m" || gender == "f" || gender == "o" {
		score += 10
	}
	if !p.DOB.CalculatedDate.IsZero() {
		score += 10
	}
	if p.Consented {
		score += 10
	}
	total += 40

	// Screening
	scTotal := 160
	if len(p.Screenings) > 0 {
		s := p.Screenings[0]
		if s.Diabetes.DiabetesType.Valid {
			score += 10
		}
		if s.Diabetes.Duration.Valid {
			score += 10
		}
		if s.MedicalHistory.Smoker.Valid {
			score += 10
		}
		if s.Medications.OnInsulin.Valid {
			score += 10
		}
		if s.Medications.TakingMedications.Valid {
			score += 10
		}
		if s.Measurements.BloodPressure.SBP.Valid {
			score += 10
		}
		if s.Measurements.BloodPressure.DBP.Valid {
			score += 10
		}
		if !s.Measurements.BloodPressure.AssessmentDate.CalculatedDate.IsZero() {
			score += 10
		}
		if s.Pathology.HbA1C.HbA1C.Valid {
			score += 10
		}
		if !s.Pathology.HbA1C.AssessmentDate.CalculatedDate.IsZero() {
			score += 10
		}
		if s.Pathology.Lipids.TotalCholesterol.Valid {
			score += 10
		}
		if !s.Pathology.Lipids.AssessmentDate.CalculatedDate.IsZero() {
			score += 10
		}
		if !s.Eyes.AssessmentDate.CalculatedDate.IsZero() {
			score += 10
		}
		// if s.Eyes.RightEye.VisualAcuity.Valid {
		// 	score += 10
		// }
		// if s.Eyes.RightEye.DRGrading.Valid {
		// 	score += 10
		// }
		// if s.Eyes.RightEye.DMEAssessment.Valid {
		// 	score += 10
		// }
		// if s.Eyes.LeftEye.VisualAcuity.Valid {
		// 	score += 10
		// }
		// if s.Eyes.LeftEye.DRGrading.Valid {
		// 	score += 10
		// }
		// if s.Eyes.LeftEye.DMEAssessment.Valid {
		// 	score += 10
		// }
		if s.Eyes.LeftEye.DilatePupil.Valid || s.Eyes.RightEye.DilatePupil.Valid {
			score += 10
		}
		if s.Referral.Referred.Valid {
			score += 10
		}
		if s.Referral.HospitalReferred.Valid {
			score += 10
		}
	}
	total += scTotal

	// Over Reading
	ovTotal := 50
	if len(p.OverReadings) > 0 {
		o := p.OverReadings[0]
		if o.Eyes.RightEye.DRGrading.Valid {
			score += 10
		}
		if o.Eyes.RightEye.DMEAssessment.Valid {
			score += 10
		}
		if o.Eyes.LeftEye.DRGrading.Valid {
			score += 10
		}
		if o.Eyes.LeftEye.DMEAssessment.Valid {
			score += 10
		}
		if o.Referral.Referred.Valid {
			score += 10
		}
	}
	total += ovTotal

	if total > 0 {
		v := float64(score) / float64(total)
		return int(math.Round(v * 100))
	}

	return 0
}
