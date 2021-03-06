package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gobuffalo/nulls"
)

/* Referral */

// ReferralScreening model
type ReferralScreening struct {
	Referred                  nulls.Bool   `json:"referred"`
	HospitalReferred          nulls.Bool   `json:"hospital_referred"`
	ReferralRefused           nulls.Bool   `json:"referral_refused"`
	ReferralReason            nulls.String `json:"referral_reason"`
	HospitalNotReferralReason nulls.String `json:"hospital_not_referral_reason"`
	Notes                     nulls.String `json:"additional_notes"`
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
