package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// CustomDate model
type CustomDate struct {
	CalculatedDate time.Time `json:"calculated_date"`
	GivenDate      time.Time `json:"given_date"`
	Calendar       string    `json:"calendar"`
}

// Value returns database driver compatible type
func (p CustomDate) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *CustomDate) Scan(src interface{}) error {
	if src == nil {
		*p = CustomDate{}
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
