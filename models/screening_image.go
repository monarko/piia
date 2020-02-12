package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gofrs/uuid"

	"github.com/monarko/piia/helpers"
)

// ScreeningImage is used by pop to map your .model.Name.Proper.Pluralize.Underscore database table to your go code.
type ScreeningImage struct {
	ID          uuid.UUID           `json:"id" db:"id"`
	CreatedAt   time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" db:"updated_at"`
	Screening   Screening           `belongs_to:"screening" json:"-"`
	ScreeningID uuid.UUID           `json:"screening_id" db:"screening_id"`
	Status      nulls.String        `json:"status" db:"status"`
	Data        helpers.PropertyMap `json:"data" db:"data"`
}

// String is not required by pop and may be deleted
func (s ScreeningImage) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// ScreeningImages is not required by pop and may be deleted
type ScreeningImages []ScreeningImage

// String is not required by pop and may be deleted
func (s ScreeningImages) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (s *ScreeningImage) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (s *ScreeningImage) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (s *ScreeningImage) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
