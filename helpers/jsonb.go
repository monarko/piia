package helpers

import (
	"database/sql/driver"
	"encoding/json"
)

// PropertyMap Type
type PropertyMap map[string]interface{}

// Value returns database driver compatible type
func (p PropertyMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *PropertyMap) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		source = []byte(`{}`)
	}

	err := json.Unmarshal(source, p)
	if err != nil {
		return err
	}

	return nil
}
