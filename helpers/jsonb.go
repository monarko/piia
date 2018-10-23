package helpers

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
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
		return errors.New("type assertion .([]byte) failed")
	}

	// var i interface{}
	err := json.Unmarshal(source, p)
	if err != nil {
		return err
	}

	// *p, ok = i.(map[string]interface{})
	// if !ok {
	// 	return errors.New("type assertion .(map[string]interface{}) failed")
	// }

	return nil
}
