package models

import (
	"time"

	"github.com/gobuffalo/pop/nulls"
)

// AnalyticsScreening object
type AnalyticsScreening struct {
	CreatedDate   time.Time    `json:"CreatedDate" db:"CreatedDate"`
	ParticipantID string       `json:"ParticipantID" db:"ParticipantID"`
	Age           int          `json:"Age" db:"Age"`
	Gender        string       `json:"Gender" db:"Gender"`
	VAOD          nulls.String `json:"VAOD" db:"VAOD"`
	VAPreviousOD  nulls.String `json:"VAPreviousOD" db:"VAPreviousOD"`
	DRGradeOD     nulls.String `json:"DRGradeOD" db:"DRGradeOD"`
	DMEOD         nulls.String `json:"DMEOD" db:"DMEOD"`
	VAOS          nulls.String `json:"VAOS" db:"VAOS"`
	VAPreviousOS  nulls.String `json:"VAPreviousOS" db:"VAPreviousOS"`
	DRGradeOS     nulls.String `json:"DRGradeOS" db:"DRGradeOS"`
	DMEOS         nulls.String `json:"DMEOS" db:"DMEOS"`
	DrReferral    nulls.String `json:"DrReferral" db:"DrReferral"`
	ReferralNotes nulls.String `json:"ReferralNotes" db:"ReferralNotes"`
}

// AnalyticsScreenings holds the list of Participant
type AnalyticsScreenings []AnalyticsScreening
