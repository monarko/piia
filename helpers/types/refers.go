package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// ReferredMessageElement model
type ReferredMessageElement struct {
	Attended             bool     `json:"attended"`
	Plans                []string `json:"plans"`
	ReferredForTreatment bool     `json:"referred_for_treatment"`
	FollowUpPlan         string   `json:"follow_up_plan"`
}

// Value returns database driver compatible type
func (p ReferredMessageElement) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *ReferredMessageElement) Scan(src interface{}) error {
	if src == nil {
		*p = ReferredMessageElement{}
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
