package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/pkg/errors"
)

// Mapping object
type Mapping map[string]interface{}

// Audit object
type Audit struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	ModelType string    `json:"model_type" db:"model_type"`
	ModelID   uuid.UUID `json:"model_id" db:"model_id"`
	OldData   Mapping   `json:"old_data" db:"old_data"`
	NewData   Mapping   `json:"new_data" db:"new_data"`
	Changes   Mapping   `json:"changes" db:"changes"`
	User      User      `belongs_to:"user" json:"user"`
	UserID    uuid.UUID `json:"-" db:"user_id"`
}

// Value returns database driver compatible type
func (p Mapping) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan converts []byte to interface{} object
func (p *Mapping) Scan(src interface{}) error {
	if src == nil {
		*p = Mapping{}
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

// String is not required by pop and may be deleted
func (a Audit) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// Audits is not required by pop and may be deleted
type Audits []Audit

// String is not required by pop and may be deleted
func (a Audits) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (a *Audit) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (a *Audit) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (a *Audit) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
