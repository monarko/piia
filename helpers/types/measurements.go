package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gobuffalo/pop/nulls"
)

/* Measurements */

// BPScreening model
type BPScreening struct {
	SBP            nulls.Int  `json:"sbp"`
	DBP            nulls.Int  `json:"dbp"`
	AssessmentDate CustomDate `json:"assessment_date"`
}

// Value returns database driver compatible type
func (p BPScreening) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *BPScreening) Scan(src interface{}) error {
	if src == nil {
		*p = BPScreening{}
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

// MeasurementScreening model
type MeasurementScreening struct {
	BloodPressure BPScreening `json:"blood_pressure"`
}

// Value returns database driver compatible type
func (p MeasurementScreening) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *MeasurementScreening) Scan(src interface{}) error {
	if src == nil {
		*p = MeasurementScreening{}
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
