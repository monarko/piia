package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gobuffalo/pop/nulls"
)

/* Diabetes */

// DiabetesScreening model
type DiabetesScreening struct {
	DiabetesType nulls.String `json:"diabetes_type"`
	Duration     nulls.Int    `json:"duration"`
	DurationType nulls.String `json:"duration_type"`
}

// Value returns database driver compatible type
func (p DiabetesScreening) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *DiabetesScreening) Scan(src interface{}) error {
	if src == nil {
		*p = DiabetesScreening{}
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
