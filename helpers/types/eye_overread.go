package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

/* Eye Assessments */

// EyeOverRead model
type EyeOverRead struct {
	DRGrading     string `json:"dr"`
	DMEAssessment string `json:"dme"`
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
