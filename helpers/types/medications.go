package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

/* Medications */

// MedicationScreening model
type MedicationScreening struct {
	TakingMedications bool `json:"taking_medications"`
	OnInsulin         bool `json:"on_insulin"`
}

// Value returns database driver compatible type
func (p MedicationScreening) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *MedicationScreening) Scan(src interface{}) error {
	if src == nil {
		*p = MedicationScreening{}
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
