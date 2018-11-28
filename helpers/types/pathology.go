package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gobuffalo/pop/nulls"
)

/* Pathology */

// HbA1CScreening model
type HbA1CScreening struct {
	HbA1C          nulls.Float64 `json:"value"`
	Unit           nulls.String  `json:"unit"`
	AssessmentDate CustomDate    `json:"assessment_date"`
}

// Value returns database driver compatible type
func (p HbA1CScreening) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *HbA1CScreening) Scan(src interface{}) error {
	if src == nil {
		*p = HbA1CScreening{}
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

// LipidScreening model
type LipidScreening struct {
	TotalCholesterol nulls.Float64 `json:"total_cholesterol"`
	HDL              nulls.Float64 `json:"hdl"`
	LDL              nulls.Float64 `json:"ldl"`
	TG               nulls.Float64 `json:"tg"`
	Unit             nulls.String  `json:"unit"`
	AssessmentDate   CustomDate    `json:"assessment_date"`
}

// Value returns database driver compatible type
func (p LipidScreening) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *LipidScreening) Scan(src interface{}) error {
	if src == nil {
		*p = LipidScreening{}
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

// PathologyScreening model
type PathologyScreening struct {
	HbA1C  HbA1CScreening `json:"hba1c"`
	Lipids LipidScreening `json:"lipids"`
}

// Value returns database driver compatible type
func (p PathologyScreening) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *PathologyScreening) Scan(src interface{}) error {
	if src == nil {
		*p = PathologyScreening{}
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
