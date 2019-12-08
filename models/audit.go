package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
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

// Audits is not required by pop and may be deleted
type Audits []Audit
