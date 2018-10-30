package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gobuffalo/pop/nulls"
)

/* Eye Assessments */

// EyeOverRead model
type EyeOverRead struct {
	DRGrading            nulls.String `json:"dr"`
	DMEAssessment        nulls.String `json:"dme"`
	SuspectedPathologies []string     `json:"suspected_pathologies"`
}

// EyeOverReading model
type EyeOverReading struct {
	RightEye EyeOverRead `json:"right"`
	LeftEye  EyeOverRead `json:"left"`
}

// Value returns database driver compatible type
func (p EyeOverReading) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *EyeOverReading) Scan(src interface{}) error {
	if src == nil {
		*p = EyeOverReading{}
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
