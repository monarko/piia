package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Permission model
type Permission struct {
	OverRead         bool `json:"overread"`
	Screening        bool `json:"screening"`
	StudyCoordinator bool `json:"study_coordinator"`
	ReferralTracker  bool `json:"referral_tracker"`
}

// Value returns database driver compatible type
func (p Permission) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *Permission) Scan(src interface{}) error {
	if src == nil {
		*p = Permission{}
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
