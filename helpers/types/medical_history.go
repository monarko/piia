package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

/* Medical History */

// MedicalHistoryScreening model
type MedicalHistoryScreening struct {
	Morbidities []string `json:"morbidities"`
}

// Value returns database driver compatible type
func (p MedicalHistoryScreening) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *MedicalHistoryScreening) Scan(src interface{}) error {
	if src == nil {
		*p = MedicalHistoryScreening{}
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
