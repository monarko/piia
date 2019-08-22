package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gobuffalo/nulls"
)

/* Eye Assessments */

// EyeAssessment model
type EyeAssessment struct {
	VisualAcuity      nulls.String `json:"visual_acuity"`
	LastVisualAcuity  nulls.String `json:"last_visual_acuity"`
	DRGrading         nulls.String `json:"dr"`
	DMEAssessment     nulls.String `json:"dme"`
	DilatePupil       nulls.Bool   `json:"dilate_pupil"`
	SuspectedCataract nulls.Bool   `json:"suspected_cataract"`
}

// EyeScreening model
type EyeScreening struct {
	RightEye       EyeAssessment `json:"right"`
	LeftEye        EyeAssessment `json:"left"`
	AssessmentDate CustomDate    `json:"assessment_date"`
}

// Value returns database driver compatible type
func (p EyeScreening) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *EyeScreening) Scan(src interface{}) error {
	if src == nil {
		*p = EyeScreening{}
		return nil
	}

	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	err := json.Unmarshal(source, p)
	if err != nil {
		return err
	}

	return nil
}
