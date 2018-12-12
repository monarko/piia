package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gobuffalo/pop/nulls"
)

/* Referral */

// ReferralScreening model
type ReferralScreening struct {
	Referred bool         `json:"referred"`
	Notes    nulls.String `json:"additional_notes"`
}

// Value returns database driver compatible type
func (p ReferralScreening) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *ReferralScreening) Scan(src interface{}) error {
	if src == nil {
		*p = ReferralScreening{}
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
